package spotifyTiles

import (
	"context"
	"fmt"
	"github.com/codemicro/platform/platform"
	"github.com/codemicro/platform/platform/util/htmlutil"
	"github.com/julienschmidt/httprouter"
	g "github.com/maragudk/gomponents"
	"github.com/maragudk/gomponents/html"
	"net/http"
	"regexp"
	"time"
)

func indexHandler(rw http.ResponseWriter, rq *http.Request, _ httprouter.Params) error {
	tok, err := loadToken()
	if err != nil {
		return fmt.Errorf("load token: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	var userInfo *user
	if tok != nil {
		userInfo, err = getUser(ctx, oauthConf.Client(ctx, tok))
		if err != nil {
			return err
		}
	}

	currentUser := "none"
	if userInfo != nil {
		if userInfo.DisplayName != "" {
			currentUser = userInfo.DisplayName + "/"
		}
		currentUser += userInfo.Id
	}

	rw.Header().Set("Content-Type", "text/html")

	var cronNodes []g.Node
	if entry := platform.GetCronEntry(recurringJobID); entry.Job != nil {
		status := "ok"
		if recurringJobHadError {
			status = "errored"
		}
		previousRunStr := entry.Prev.Format(time.DateTime)
		if entry.Prev.IsZero() {
			previousRunStr = "never"
		}
		cronNodes = append(cronNodes, html.Li(g.Textf("Job last run: %s (status %s)", previousRunStr, status)))
		cronNodes = append(cronNodes, html.Li(g.Textf("Job next run: %s (in %s)", entry.Next.Format(time.DateTime), entry.Next.Sub(time.Now()).Truncate(time.Second).String())))
	}

	return htmlutil.BasePage(
		"Spotify tiles",
		html.H1(g.Text("Spotify Tiles")),
		html.A(g.Text("[See detected playlists]"), g.Attr("href", "/playlists")),
		html.Ul(
			html.Li(g.Textf("Current user: %s ", currentUser), html.A(g.Attr("href", "/oauth/outbound"), g.Text("[auth]"))),
			html.Li(g.Text("OAuth redirect URL: "), html.Code(g.Text(deriveOauthRedirectURL(rq)))),
			g.If(len(cronNodes) != 0, g.Group(cronNodes)),
		),
	).Render(rw)
}

var playlistNameRegexp = regexp.MustCompile(`([A-Za-z' ]+ \()?(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) \d{4}\)?`)

func detectedPlaylistsHandler(rw http.ResponseWriter, _ *http.Request, _ httprouter.Params) error {
	tok, err := loadToken()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	client := oauthConf.Client(ctx, tok)

	userInfo, err := getUser(ctx, client)
	if err != nil {
		return err
	}

	playlists, err := getPlaylists(ctx, client, userInfo.Id)
	if err != nil {
		return err
	}

	{
		n := 0
		for _, x := range playlists {
			if playlistNameRegexp.MatchString(x.Name) {
				playlists[n] = x
				n += 1
			}
		}
		playlists = playlists[:n]
	}

	return htmlutil.BasePage(
		"Spotify tiles",
		html.H1(g.Text("Spotify Tiles")),
		html.A(g.Text("[Back to index]"), g.Attr("href", "/")),
		html.H2(g.Text("Detected Playlists")),
		html.Ul(g.Map(playlists, func(p *playlist) g.Node {
			return html.Li(g.Text(p.Name + " "))
		})...),
	).Render(rw)
}

func tileHandler(rw http.ResponseWriter, _ *http.Request, _ httprouter.Params) error {
	if err := runTilesTask(); err != nil {
		return err
	}

	rw.WriteHeader(204)
	return nil
}
