package readingList

import (
	"github.com/codemicro/platform/platform"
	"github.com/codemicro/platform/platform/storage"
	"github.com/codemicro/platform/platform/util"
	"github.com/go-playground/validator"
	"github.com/julienschmidt/httprouter"
)

const moduleName = "readinglist"

func init() {
	router.GET("/", util.WrapHandler(indexHandler))
	router.GET("/csv", util.WrapHandler(sourceCSVHandler))
	router.GET("/map", util.WrapHandler(mapHandler))
	router.GET("/api/add", util.WrapHandler(addHandler))

	platform.RegisterProvider(moduleName, router)
}

var (
	router = httprouter.New()
	store  = util.Must(storage.New(moduleName))
)

var validate = validator.New()

type inputs struct {
	URL         string `validate:"required,url"`
	Title       string `validate:"required"`
	Description string
	Image       string
}

func (i *inputs) Validate() error {
	return validate.Struct(i)
}
