package readingList

import (
	"context"
	"fmt"
	"github.com/carlmjohnson/requests"
	"github.com/codemicro/platform/config"
	"github.com/codemicro/platform/platform/util/htmlutil"
	"github.com/go-playground/validator"
	"github.com/julienschmidt/httprouter"
	g "github.com/maragudk/gomponents"
	"github.com/maragudk/gomponents/html"
	"net/http"
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

	return htmlutil.BasePage("Success!", html.Span(
		html.StyleAttr("color: darkgreen;"),
		g.Text("Success!"),
	),
		html.Script(g.Rawf(`setTimeout(function(){window.location.replace(%#v);}, 500);`, data.NextURL)),
	).Render(rw)
}
