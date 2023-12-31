package config

import (
	"fmt"
	"git.tdpain.net/codemicro/kindle-dashboard/imagegen"
	"git.tdpain.net/pkg/cfger"
	"log/slog"
	"os"
	"strings"
	"sync"
)

type HTTP struct {
	Host string
	Port int
}

func (h *HTTP) Address() string {
	return fmt.Sprintf("%s:%d", h.Host, h.Port)
}

type ReadingList struct {
	Token             string
	GithubAccessToken string
}

type SpotifyTiles struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type Config struct {
	Debug           bool
	HTTP            *HTTP
	ReadingList     *ReadingList
	SpotifyTiles    *SpotifyTiles
	KindleDashboard *imagegen.Config
	HostSuffix      string
}

var (
	conf     *Config
	loadOnce = new(sync.Once)
)

func Get() *Config {
	var outerErr error
	loadOnce.Do(func() {
		cl := cfger.New()
		if err := cl.Load("config.yml"); err != nil {
			outerErr = err
			return
		}

		conf = &Config{
			Debug: cl.WithDefault("debug", false).AsBool(),
			HTTP: &HTTP{
				Host: cl.WithDefault("http.host", "127.0.0.1").AsString(),
				Port: cl.WithDefault("http.port", 8080).AsInt(),
			},
			ReadingList: &ReadingList{
				Token:             cl.Required("readingList.token").AsString(),
				GithubAccessToken: cl.Required("readingList.githubAccessToken").AsString(),
			},
			SpotifyTiles: &SpotifyTiles{
				ClientID:     cl.Required("spotifyTiles.clientID").AsString(),
				ClientSecret: cl.Required("spotifyTiles.clientSecret").AsString(),
				RedirectURI:  cl.Get("spotifyTiles.redirectURI").AsString(),
			},
			KindleDashboard: &imagegen.Config{
				WxLocation:               cl.Required("kindle.weatherLocationID").AsString(),
				MetOfficeDatapointAPIKey: cl.Required("kindle.metOfficeDatapointToken").AsString(),
				TrainsLocation:           cl.Required("kindle.trainsLocationCRS").AsString(),
				RTTUsername:              cl.Required("kindle.rttUsername").AsString(),
				RTTPassword:              cl.Required("kindle.rttPassword").AsString(),
			},
			HostSuffix: cl.Required("hostSuffix").AsString(),
		}

		if !strings.HasPrefix(conf.HostSuffix, ".") {
			conf.HostSuffix = "." + conf.HostSuffix
		}
	})

	if outerErr != nil {
		slog.Error("fatal error when loading configuration", "err", outerErr)
		os.Exit(1)
	}

	return conf
}
