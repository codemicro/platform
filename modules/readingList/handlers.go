package readingList

import (
	"context"
	"errors"
	"fmt"
	"github.com/carlmjohnson/requests"
	"github.com/codemicro/platform/config"
	"github.com/codemicro/platform/platform/util/htmlutil"
	"github.com/go-playground/validator"
	"github.com/julienschmidt/httprouter"
	g "github.com/maragudk/gomponents"
	"github.com/maragudk/gomponents/html"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func addHandler(rw http.ResponseWriter, rq *http.Request, _ httprouter.Params) error {
	rw.Header().Set("Access-Control-Allow-Headers", "content-type,authorization")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Content-Type", "text/html")

	data := &struct {
		inputs
		NextURL string `validate:"required,url"`
		Token   string `validate:"required"`
	}{
		inputs: inputs{
			URL:         rq.URL.Query().Get("url"),
			Title:       rq.URL.Query().Get("title"),
			Description: rq.URL.Query().Get("description"),
			Image:       rq.URL.Query().Get("image"),
		},
		NextURL: rq.URL.Query().Get("nexturl"),
		Token:   rq.URL.Query().Get("token"),
	}

	{
		validate := validator.New()
		err := validate.Struct(data)
		if err != nil {
			rw.WriteHeader(400)
			n := htmlutil.BasePage("Bad request", g.Text("Bad request"), html.Br(), htmlutil.UnorderedList(strings.Split(err.Error(), "\n")))
			return n.Render(rw)
		}
	}

	if data.Token != config.Get().ReadingList.Token {
		rw.WriteHeader(401)
		n := htmlutil.BasePage("Invalid token", g.Text("Unauthorised - invalid token"))
		return n.Render(rw)
	}

	bodyData := &struct {
		EventType     string `json:"event_type"`
		ClientPayload any    `json:"client_payload"`
	}{
		EventType:     "rl-append",
		ClientPayload: data.inputs,
	}

	rctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	// trigger workflow in repository
	err := requests.
		URL("https://api.github.com/repos/codemicro/readingList/dispatches").
		Header("Accept", "application/vnd.github+json").
		Header("Authorization", "token "+config.Get().ReadingList.GithubAccessToken).
		BodyJSON(bodyData).
		Fetch(rctx)
	if err != nil {
		return err
	}

	if err := addRowToCSV(&data.inputs); err != nil {
		return fmt.Errorf("add row to CSV: %w", err)
	}

	if err := generateMapFile(); err != nil {
		return fmt.Errorf("generate map file: %w", err)
	}

	return htmlutil.BasePage("Success!", html.Span(
		html.StyleAttr("color: darkgreen;"),
		g.Text("Success!"),
	),
		html.Script(g.Rawf(`setTimeout(function(){window.location.replace(%#v);}, 500);`, data.NextURL)),
	).Render(rw)
}

func indexHandler(rw http.ResponseWriter, _ *http.Request, _ httprouter.Params) error {
	return htmlutil.BasePage(
		"Reading list",
		html.H1(g.Text("Reading list")),
		html.A(g.Text("Source CSV"), g.Attr("href", "/csv")),
	).Render(rw)
}

func sourceCSVHandler(rw http.ResponseWriter, rq *http.Request, _ httprouter.Params) error {
	f, err := openReadingListFile()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			rw.WriteHeader(404)
			_, _ = rw.Write([]byte("Not found"))
			return nil
		}
		return err
	}
	defer f.Close()

	rw.Header().Set("Content-Type", "text/csv")

	rangeString := rq.Header.Get("Range")
	sp := strings.Split(rangeString, "=")
	if len(sp) == 2 && sp[0] == "bytes" {
		offsetStrings := strings.Split(sp[1], "-")
		start, err := strconv.ParseInt(offsetStrings[0], 10, 64)
		if err != nil {
			goto fullFile
		}
		_, err = f.Seek(start, 0)
		if err != nil {
			goto fullFile
		}
		if len(offsetStrings) > 1 {
			end, err := strconv.ParseInt(offsetStrings[1], 10, 64)
			if err != nil || end < start {
				_, _ = f.Seek(0, 0)
				goto fullFile
			}
			rw.WriteHeader(http.StatusPartialContent)
			_, err = io.CopyN(rw, f, end-start)
			return err
		}
		rw.WriteHeader(http.StatusPartialContent)
	}

fullFile:
	_, err = io.Copy(rw, f)
	return err
}

func mapHandler(rw http.ResponseWriter, _ *http.Request, _ httprouter.Params) error {
	f, err := os.Open(store.MakePath(mapFilename))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			rw.WriteHeader(404)
			_, _ = rw.Write([]byte("Not found"))
			return nil
		}
		return err
	}
	defer f.Close()

	rw.Header().Set("Content-Type", "application/json")

	_, err = io.Copy(rw, f)
	return err
}
