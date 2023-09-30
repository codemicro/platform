package readingList

import (
	"context"
	"github.com/carlmjohnson/requests"
	"github.com/codemicro/platform/config"
	"github.com/codemicro/platform/platform"
	"github.com/codemicro/platform/platform/util"
	"github.com/codemicro/platform/platform/util/htmlutil"
	rltransport "github.com/codemicro/readingList/transport"
	"github.com/go-playground/validator"
	"github.com/julienschmidt/httprouter"
	g "github.com/maragudk/gomponents"
	"github.com/maragudk/gomponents/html"
	"net/http"
	"strings"
	"time"
)

func init() {
	router.GET("/api/add", util.WrapHandler(addHandler))

	platform.RegisterProvider("readinglist", router)
}

var router = httprouter.New()

func addHandler(rw http.ResponseWriter, rq *http.Request, _ httprouter.Params) error {
	rw.Header().Set("Access-Control-Allow-Headers", "content-type,authorization")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Content-Type", "text/html")

	data := &struct {
		rltransport.Inputs
		NextURL string `validate:"required,url" query:"nexturl"`
		Token   string `validate:"required" query:"token"`
	}{
		Inputs: rltransport.Inputs{
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
		ClientPayload: data.Inputs,
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

	return htmlutil.BasePage("Success!", html.Span(
		html.StyleAttr("color: darkgreen;"),
		g.Text("Success!"),
	),
		html.Script(g.Rawf(`setTimeout(function(){window.location.replace(%#v);}, 500);`, data.NextURL)),
	).Render(rw)
}
