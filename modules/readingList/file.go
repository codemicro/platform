package readingList

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const readingListFilename = "readingList.csv"

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
		doesCSVExist = err != nil && errors.Is(err, os.ErrNotExist)
	}

	fileFlags := os.O_APPEND | os.O_WRONLY

	var records [][]string

	if !doesCSVExist {
		fileFlags = fileFlags | os.O_CREATE
		records = append(records, []string{"url", "title", "description", "image", "date", "hnurl"})
	}

	timeStr, _ := time.Now().MarshalText()
	records = append(records, []string{data.URL, data.Title, data.Description, string(timeStr), hnURL})

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
