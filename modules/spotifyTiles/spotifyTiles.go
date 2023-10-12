package spotifyTiles

import (
	"encoding/json"
	"errors"
	"github.com/codemicro/platform/platform"
	"github.com/codemicro/platform/platform/storage"
	"github.com/codemicro/platform/platform/util"
	"github.com/julienschmidt/httprouter"
	"os"
)

const moduleName = "spotifytiles"

func init() {
	//router.GET("/", util.WrapHandler(indexHandler))

	router.GET("/", util.WrapHandler(indexHandler))
	router.GET("/playlists", util.WrapHandler(detectedPlaylistsHandler))
	router.GET("/oauth/outbound", util.WrapHandler(oauthOutbound))
	router.GET("/oauth/inbound", util.WrapHandler(oauthInbound))
	router.GET("/tile", util.WrapHandler(tileHandler))

	platform.RegisterProvider(moduleName, router)
}

var (
	router               = httprouter.New()
	store                = util.Must(storage.New(moduleName))
	recurringJobHadError = false
	recurringJobID       = util.Must(platform.RegisterRecurringTask("*/30 * * * *", func() error {
		err := runTilesTask()
		recurringJobHadError = err != nil
		return err
	}))
)

func dumpToDisk[T any](m T, fname string) error {
	jdat, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if err := os.WriteFile(fname, jdat, 0644); err != nil {
		return err
	}
	return nil
}

func restoreFromDisk[T any](fname string) (T, error) {
	var tok T
	rawdat, err := os.ReadFile(fname)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return tok, nil
		}
		return tok, err
	}
	if err := json.Unmarshal(rawdat, &tok); err != nil {
		return tok, err
	}
	return tok, nil
}
