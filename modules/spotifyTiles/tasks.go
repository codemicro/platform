package spotifyTiles

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

func runTilesTask() error {
	tok, err := loadToken()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
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

	knownSnapshots, err := loadKnownSnapshots()
	if err != nil {
		return err
	}

	for _, x := range playlists {
		if !playlistNameRegexp.MatchString(x.Name) {
			continue
		}

		if ksid, found := knownSnapshots[x.Id]; !found {
			// new playlist
			knownSnapshots[x.Id] = x.SnapshotId
		} else if ksid != x.SnapshotId {
			// updated playlist, regenerate
			slog.Debug("playlists updated", "id", x.Id, "ssid", x.SnapshotId)
			knownSnapshots[x.Id] = x.SnapshotId
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
			imgs, err := getPlaylistAlbumImages(ctx, client, x.Id)
			defer cancel()
			if err != nil {
				return err
			}
			img, err := generateFromAlbumImages(imgs)
			if err != nil {
				if !errors.Is(err, errTooFewImages) {
					return fmt.Errorf("generate playlist image for %s: %w", x.Id, err)
				}
			} else {
				if err := setPlaylistImage(ctx, client, x.Id, img); err != nil {
					return fmt.Errorf("set playlist image: %w", err)
				}
			}
		}
	}

	if err := saveKnownSnapshots(knownSnapshots); err != nil {
		return err
	}

	return nil
}
