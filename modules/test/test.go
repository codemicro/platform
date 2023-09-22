package test

import (
	"fmt"
	"github.com/codemicro/platform/platform"
	"github.com/codemicro/platform/platform/util"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strings"
)

func init() {
	router.GET("/hello", util.WrapHandler(sayHello))
	platform.RegisterProvider("test", router)
}

var router = httprouter.New()

func sayHello(rw http.ResponseWriter, rq *http.Request, _ httprouter.Params) error {
	var sb strings.Builder

	sb.WriteString("Hello! You've reached this function.\n\n")
	sb.WriteString(fmt.Sprintf("%#v", *rq))

	_, _ = rw.Write([]byte(sb.String()))

	return nil
}
