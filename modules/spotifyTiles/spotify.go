package spotifyTiles

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/carlmjohnson/requests"
	"image"
	"image/jpeg"
	"net/http"
)

type paginated[T any] struct {
	Href     string `json:"href"`
	Limit    int    `json:"limit"`
	Next     string `json:"next"`
	Offset   int    `json:"offset"`
	Previous string `json:"previous"`
	Total    int    `json:"total"`
	Items    []T    `json:"items"`
}

type playlist struct {
	Collaborative bool   `json:"collaborative"`
	Description   string `json:"description"`
	ExternalUrls  struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Href   string `json:"href"`
	Id     string `json:"id"`
	Images []struct {
		Height int    `json:"height"`
		Url    string `json:"url"`
		Width  int    `json:"width"`
	} `json:"images"`
	Name  string `json:"name"`
	Owner struct {
		DisplayName  string `json:"display_name"`
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href string `json:"href"`
		Id   string `json:"id"`
		Type string `json:"type"`
		Uri  string `json:"uri"`
	} `json:"owner"`
	PrimaryColor interface{} `json:"primary_color"`
	Public       bool        `json:"public"`
	SnapshotId   string      `json:"snapshot_id"`
	Tracks       struct {
		Href  string `json:"href"`
		Total int    `json:"total"`
	} `json:"tracks"`
	Type string `json:"type"`
	Uri  string `json:"uri"`
}

type albumImages struct {
	Id     string `json:"id"`
	Images []struct {
		Url    string `json:"url"`
		Height int    `json:"height"`
		Width  int    `json:"width"`
	} `json:"images"`
}

type user struct {
	DisplayName  string `json:"display_name"`
	ExternalUrls struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Href   string `json:"href"`
	Id     string `json:"id"`
	Images []struct {
		Url    string `json:"url"`
		Height int    `json:"height"`
		Width  int    `json:"width"`
	} `json:"images"`
	Type      string `json:"type"`
	Uri       string `json:"uri"`
	Followers struct {
		Href  string `json:"href"`
		Total int    `json:"total"`
	} `json:"followers"`
}

func getUser(ctx context.Context, client *http.Client) (*user, error) {
	var dat *user
	err := requests.
		URL("https://api.spotify.com").
		Pathf("/v1/me").
		ToJSON(&dat).
		Client(client).
		Fetch(ctx)

	if err != nil {
		return nil, fmt.Errorf("get user information: %w", err)
	}

	return dat, nil
}

func getAllPaginated[T any](ctx context.Context, client *http.Client, url string) ([]T, error) {
	var res []T
	for url != "" {
		var dat paginated[T]
		err := requests.
			URL(url).
			ToJSON(&dat).
			Client(client).
			Fetch(ctx)
		if err != nil {
			return nil, err
		}
		url = dat.Next
		res = append(res, dat.Items...)
	}
	return res, nil
}

func getPlaylists(ctx context.Context, client *http.Client, userID string) ([]*playlist, error) {
	p, err := getAllPaginated[*playlist](ctx, client, fmt.Sprintf("https://api.spotify.com/v1/users/%s/playlists?limit=50", userID))
	if err != nil {
		return nil, fmt.Errorf("get all playlists: %w", err)
	}
	return p, nil
}

func getPlaylistAlbumImages(ctx context.Context, client *http.Client, playlistID string) ([]*albumImages, error) {
	type wrapper struct {
		Track struct {
			Album *albumImages `json:"album"`
		} `json:"track"`
	}
	p, err := getAllPaginated[*wrapper](ctx, client, fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks?fields=next%%2Citems.track.album%%28id%%2Cimages%%29&limit=50", playlistID))
	if err != nil {
		return nil, fmt.Errorf("get all playlists: %w", err)
	}
	var res []*albumImages
	seen := make(map[string]struct{})
	for _, x := range p {
		if _, found := seen[x.Track.Album.Id]; !found && len(x.Track.Album.Images) != 0 {
			var ok bool
			for _, y := range x.Track.Album.Images {
				ok = ok || y.Height == y.Width
			}
			if ok {
				seen[x.Track.Album.Id] = struct{}{}
				res = append(res, x.Track.Album)
			}
		}
	}
	return res, nil
}

func setPlaylistImage(ctx context.Context, client *http.Client, playlistID string, img image.Image) error {
	buf := new(bytes.Buffer)
	enc := base64.NewEncoder(base64.StdEncoding, buf)
	if err := jpeg.Encode(enc, img, &jpeg.Options{Quality: 75}); err != nil {
		return err
	}
	if err := enc.Close(); err != nil {
		return err
	}

	err := requests.
		URL("https://api.spotify.com").
		Pathf("/v1/playlists/%s/images", playlistID).
		BodyReader(buf).
		ContentType("image/jpeg").
		Method(http.MethodPut).
		Client(client).
		Fetch(ctx)

	return err
}

type snapshots map[string]string

func saveKnownSnapshots(s snapshots) error {
	if err := dumpToDisk(s, store.MakePath("snapshots.json")); err != nil {
		return fmt.Errorf("save snapshots: %w", err)
	}
	return nil
}

func loadKnownSnapshots() (snapshots, error) {
	s, err := restoreFromDisk[snapshots](store.MakePath("snapshots.json"))
	if err != nil {
		return nil, fmt.Errorf("load snapshots: %w", err)
	}
	if s == nil {
		return make(snapshots), nil
	}
	return s, nil
}
