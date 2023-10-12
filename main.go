package main

import (
	"github.com/codemicro/platform/config"
	"github.com/codemicro/platform/platform"
	"log/slog"
	"os"

	_ "github.com/codemicro/platform/modules/avatars"
	_ "github.com/codemicro/platform/modules/readingList"
	_ "github.com/codemicro/platform/modules/spotifyTiles"
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
	if conf.Debug {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
	}
	platform.StartCron()
	return platform.ListenAndServe(conf.HTTP.Address())
}
