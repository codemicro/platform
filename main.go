package main

import (
	"github.com/codemicro/platform/config"
	"github.com/codemicro/platform/platform"
	"log/slog"
	"os"

	_ "github.com/codemicro/platform/modules/test"
)

func main() {
	if err := run(); err != nil {
		slog.Error("unhandled error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	conf := config.Get()
	slog.Info("HTTP alive!", "address", conf.HTTP.Address())
	return platform.ListenAndServe(conf.HTTP.Address())
}
