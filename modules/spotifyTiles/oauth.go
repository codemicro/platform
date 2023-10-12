package spotifyTiles

import (
	"context"
	"fmt"
	"github.com/codemicro/platform/config"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
	"time"
)

var oauthConf = &oauth2.Config{
	ClientID:     config.Get().SpotifyTiles.ClientID,
	ClientSecret: config.Get().SpotifyTiles.ClientSecret,
	Scopes:       []string{"playlist-read-private", "playlist-modify-public", "playlist-modify-private", "ugc-image-upload"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://accounts.spotify.com/authorize",
		TokenURL: "https://accounts.spotify.com/api/token",
	},
}

func oauthOutbound(rw http.ResponseWriter, rq *http.Request, _ httprouter.Params) error {
	url := oauthConf.AuthCodeURL("na", oauth2.SetAuthURLParam("redirect_uri", deriveOauthRedirectURL(rq)), oauth2.SetAuthURLParam("show_dialog", "true"))
	rw.Header().Set("Location", url)
	rw.WriteHeader(301)
	return nil
}

func oauthInbound(rw http.ResponseWriter, rq *http.Request, _ httprouter.Params) error {
	code := rq.URL.Query().Get("code")
	if code == "" {
		rw.WriteHeader(400)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	tok, err := oauthConf.Exchange(ctx, code,
		oauth2.SetAuthURLParam("grant_type", "authorization_code"),
		oauth2.SetAuthURLParam("redirect_uri", deriveOauthRedirectURL(rq)))
	if err != nil {
		return fmt.Errorf("oauth2 exchange: %w", err)
	}

	if err := saveToken(tok); err != nil {
		return fmt.Errorf("save oauth2 token: %w", err)
	}

	rw.WriteHeader(301)
	rw.Header().Set("Location", "/")
	return nil
}

func deriveOauthRedirectURL(rq *http.Request) string {
	redirectURI := config.Get().SpotifyTiles.RedirectURI
	if redirectURI == "" {
		redirectURI = (&url.URL{
			Scheme: "https",
			Host:   rq.Host,
			Path:   "/oauth/inbound",
		}).String()
	}
	return redirectURI
}

func saveToken(tok *oauth2.Token) error {
	if err := dumpToDisk(tok, store.MakePath("spotifyCredentials.json")); err != nil {
		return fmt.Errorf("save token: %w", err)
	}
	return nil
}

func loadToken() (*oauth2.Token, error) {
	tok, err := restoreFromDisk[*oauth2.Token](store.MakePath("spotifyCredentials.json"))
	if err != nil {
		return nil, fmt.Errorf("load token: %w", err)
	}
	return tok, nil
}
