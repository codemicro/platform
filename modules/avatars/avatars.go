package avatars

import (
	"github.com/codemicro/platform/platform"
	"github.com/codemicro/platform/platform/util"
	"github.com/jakobvarmose/go-qidenticon"
	"github.com/julienschmidt/httprouter"
	"image/png"
	"net/http"
	"strconv"
)

func init() {
	router.GET("/:data", util.WrapHandler(handler))

	platform.RegisterProvider("avatars", router)
}

var router = httprouter.New()

func handler(rw http.ResponseWriter, rq *http.Request, params httprouter.Params) error {
	dat := params.ByName("data")
	size, err := strconv.Atoi(rq.URL.Query().Get("size"))
	if err != nil {
		size = 200
	}
	img := qidenticon.Render(qidenticon.Code(dat), size, qidenticon.DefaultSettings())
	rw.Header().Set("Content-Type", "image/png")
	return png.Encode(rw, img)
}
