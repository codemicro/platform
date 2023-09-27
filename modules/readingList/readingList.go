package readingList

import (
	"context"
	"fmt"
	"github.com/carlmjohnson/requests"
	"github.com/codemicro/platform/config"
	"github.com/codemicro/platform/platform"
	"github.com/codemicro/platform/platform/util"
	rltransport "github.com/codemicro/readingList/transport"
	"github.com/go-playground/validator"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"time"
)

func init() {
	router.GET("/api/add", util.WrapHandler(addHandler))

	platform.RegisterProvider("readingList", router)
}

var router = httprouter.New()

func addHandler(rw http.ResponseWriter, rq *http.Request, _ httprouter.Params) error {
	rw.Header().Set("Access-Control-Allow-Headers", "content-type,authorization")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

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
			_, _ = rw.Write([]byte("Bad request - " + err.Error()))
			return nil
		}
	}

	if data.Token != config.Get().ReadingList.Token {
		rw.WriteHeader(401)
		_, _ = rw.Write([]byte("Invalid token"))
		return nil
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

	rw.Header().Set("Content-Type", "text/html")
	_, _ = rw.Write([]byte(fmt.Sprintf("<span style='color: darkgreen;'>Success!</span><script>setTimeout(function(){window.location.replace(%#v);}, 500);</script>", data.NextURL)))
	return nil
}
