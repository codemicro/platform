package platform

import (
	"github.com/codemicro/platform/config"
	"log/slog"
	"net/http"
	"strings"
)

type provider struct {
	Host    string
	Handler http.Handler
}

var providers []*provider

func RegisterProvider(hostPrefix string, handler http.Handler) {
	p := &provider{
		Host:    strings.ToLower(hostPrefix) + config.Get().HostSuffix,
		Handler: handler,
	}
	providers = append(providers, p)
	slog.Info("Registered new provider", "host", p.Host)
}

func ListenAndServe(address string) error {
	slog.Info("HTTP alive!", "address", address)
	return http.ListenAndServe(address, http.HandlerFunc(handler))
}

func handler(rw http.ResponseWriter, rq *http.Request) {
	host := strings.ToLower(rq.Host)
	hostWasTransformed := host != rq.Host
	for _, provider := range providers {
		if host == provider.Host {
			if hostWasTransformed {
				newURL := *rq.URL
				newURL.Host = host
				rw.Header().Set("Location", newURL.String())
				rw.WriteHeader(http.StatusFound)
				return
			}
			provider.Handler.ServeHTTP(rw, rq)
			return
		}
	}
	rw.WriteHeader(404)
}
