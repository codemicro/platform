package avatars

import (
	"crypto/sha1"
	"fmt"
	"git.tdpain.net/codemicro/kindle-dashboard/imagegen"
	"github.com/codemicro/platform/config"
	"github.com/codemicro/platform/platform"
	"github.com/codemicro/platform/platform/util"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func init() {
	router.GET("/image", util.WrapHandler(imageHandler))

	platform.RegisterProvider("kindle", router)
}

var router = httprouter.New()

func imageHandler(rw http.ResponseWriter, rq *http.Request, _ httprouter.Params) error {
	imageData, err := imagegen.Generate(config.Get().KindleDashboard)
	if err != nil {
		return err
	}

	etag := fmt.Sprintf("\"%x\"", sha1.Sum(imageData))

	if rq.Header.Get("If-None-Match") == etag {
		rw.WriteHeader(304)
		return nil
	}

	h := rw.Header()
	h.Set("ETag", etag)
	h.Set("Content-Type", "image/png")
	rw.WriteHeader(200)
	_, err = rw.Write(imageData)
	return err
}
