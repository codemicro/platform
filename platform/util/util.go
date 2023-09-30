package util

import (
	"github.com/julienschmidt/httprouter"
	"log/slog"
	"net/http"
)

func RespondStatus(rw http.ResponseWriter, status int) {
	rw.WriteHeader(status)
}

func Respond404(rw http.ResponseWriter) {
	RespondStatus(rw, 404)
}

type HandlerFuncWithError = func(http.ResponseWriter, *http.Request, httprouter.Params) error

func WrapHandler(hf HandlerFuncWithError) httprouter.Handle {
	return func(rw http.ResponseWriter, rq *http.Request, params httprouter.Params) {
		if err := hf(rw, rq, params); err != nil {
			slog.Error("unhandled error in handler", "err", err)
			rw.WriteHeader(500)
		}
	}
}

func Must[T any](x T, y error) T {
	if y != nil {
		panic(y)
	}
	return x
}
