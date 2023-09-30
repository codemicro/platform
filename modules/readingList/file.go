package readingList

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"time"
)

const readingListFilename = "readingList.csv"
const mapFilename = "readingList.map.json"

func addRowToCSV(data *inputs) error {
	if err := data.Validate(); err != nil {
		return err
	}

	hnURL, err := queryHackerNews(data.URL)
	if err != nil {
		slog.Warn("Unable to search Hacker News", "url", data.URL)
	}

	// if CSV file does not exist, create it with a header
	csvFilePath := store.MakePath(readingListFilename)

	var doesCSVExist bool
	{
		_, err := os.Stat(csvFilePath)
		doesCSVExist = err == nil
	}

	fileFlags := os.O_APPEND | os.O_WRONLY

	var records [][]string

	if !doesCSVExist {
		fileFlags = fileFlags | os.O_CREATE
		records = append(records, []string{"url", "title", "description", "image", "date", "hnurl"})
	}

	timeStr, _ := time.Now().MarshalText()
	records = append(records, []string{data.URL, data.Title, data.Description, data.Image, string(timeStr), hnURL})

	// make changes to CSV file

	f, err := os.OpenFile(csvFilePath, fileFlags, 0644)
	if err != nil {
		return err
	}

	w := csv.NewWriter(f)
	_ = w.WriteAll(records)

	if err := f.Close(); err != nil {
		return err
	}

	return nil
}

var hnHTTPClient = new(http.Client)

type hackerNewsEntry struct {
	ObjectID string `json:"objectID"`
	Points   int    `json:"points"`
}

var hackerNewsSubmissionURL = "https://news.ycombinator.com/item?id=%s"

func queryHackerNews(url string) (string, error) {

	req, err := http.NewRequest("GET", "https://hn.algolia.com/api/v1/search", nil)
	if err != nil {
		return "", err
	}

	// why does this fel so hacky
	queryParams := req.URL.Query()
	queryParams.Add("restrictSearchableAttributes", "url")
	queryParams.Add("hitsPerPage", "1000")
	queryParams.Add("query", url)
	req.URL.RawQuery = queryParams.Encode()

	resp, err := hnHTTPClient.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HN Search returned a non-200 status code: %d", resp.StatusCode)
	}

	responseBody, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var x struct {
		Hits []*hackerNewsEntry `json:"hits"`
	}

	err = json.Unmarshal(responseBody, &x)
	if err != nil {
		return "", err
	}

	var targetSubmission *hackerNewsEntry
	if len(x.Hits) == 0 {
		return "", nil
	} else if len(x.Hits) == 1 {
		targetSubmission = x.Hits[0]
	} else {
		// must be more than one hit
		var topRatedSubmission *hackerNewsEntry
		for _, entry := range x.Hits {
			if topRatedSubmission == nil || entry.Points > topRatedSubmission.Points {
				topRatedSubmission = entry
			}
		}
		targetSubmission = topRatedSubmission
	}

	return fmt.Sprintf(hackerNewsSubmissionURL, targetSubmission.ObjectID), nil
}

// openReadingListFile opens the reading list CSV for reading.
func openReadingListFile() (*os.File, error) {
	csvFilePath := store.MakePath(readingListFilename)
	f, err := os.Open(csvFilePath)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func generateMapFile() error {
	f, err := openReadingListFile()
	if err != nil {
		return fmt.Errorf("open reading list CSV: %w", err)
	}
	defer f.Close()

	// This entire function presumes that the dates contained in the file are in a strictly increasing order the further
	// into the file you get.

	type fileRange struct {
		Start int64 `json:"start"`
		End   int64 `json:"end"`
	}

	offsets := make(map[[2]int]*fileRange)

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1 // disable record length checking

	_, _ = reader.Read() // ignore the header line

	lineStart := reader.InputOffset()

	for {
		var stop bool
		record, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				stop = true
			} else {
				return fmt.Errorf("read record: %w", err)
			}
		}

		if len(record) == 0 {
			break
		}

		lineEnd := reader.InputOffset()

		dateString := record[4]
		recordTime := &time.Time{}
		if err := recordTime.UnmarshalText([]byte(dateString)); err != nil {
			return fmt.Errorf("unmarshal time: %w", err)
		}

		key := [2]int{int(recordTime.Month()), recordTime.Year()}

		fr := offsets[key]
		if fr == nil {
			fr = &fileRange{}
			offsets[key] = fr
			fr.Start = lineStart
		}

		fr.End = lineEnd - 1

		if stop {
			break
		}

		lineStart = lineEnd
	}

	var months [][2]int
	for k := range offsets {
		months = append(months, k)
	}

	sort.Slice(months, func(i, j int) bool {
		im, jm := months[i], months[j]
		if im[1] != jm[1] {
			return im[1] < jm[1]
		}
		return im[0] < jm[0]
	})

	type resp struct {
		Name  string     `json:"name"`
		Range *fileRange `json:"range"`
	}

	var res []*resp

	for _, k := range months {
		res = append(res, &resp{
			Name:  fmt.Sprintf("%s %d", time.Month(k[0]).String()[0:3], k[1]),
			Range: offsets[k],
		})
	}

	j, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	if err := os.WriteFile(store.MakePath(mapFilename), j, 0644); err != nil {
		return fmt.Errorf("dump to file: %w", err)
	}

	return nil
}
