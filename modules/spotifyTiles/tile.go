package spotifyTiles

import (
	"bytes"
	"context"
	"errors"
	"github.com/carlmjohnson/requests"
	"golang.org/x/image/draw"
	"image"
	"image/jpeg"
	"os"
	"time"
)

var errTooFewImages = errors.New("too few images")

const outputSize = 300

func generateFromAlbumImages(imgs []*albumImages) error {
	if len(imgs) <= 3 {
		return errTooFewImages
	}

	targetDim := 2
	for {
		if x := targetDim + 1; x*x > len(imgs) {
			break
		} else {
			targetDim = x
		}
	}

	//img := image.NewRGBA(image.Rect(0, 0, outputSize, outputSize))

	res, err := getImageFromURL(imgs[0].Images[0].Url)
	if err != nil {
		return err
	}

	dst := image.NewRGBA(image.Rect(0, 0, 50, 50))
	draw.ApproxBiLinear.Scale(dst, dst.Rect, res, res.Bounds(), draw.Over, nil)

	f, err := os.OpenFile("bananas.jpg", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	if err := jpeg.Encode(f, dst, nil); err != nil {
		return err
	}
	return f.Close()
	//return nil
}

func getImageFromURL(url string) (image.Image, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	buf := new(bytes.Buffer)
	err := requests.
		URL(url).
		ToBytesBuffer(buf).
		Fetch(ctx)
	if err != nil {
		return nil, err
	}
	img, err := jpeg.Decode(buf)
	if err != nil {
		return nil, err
	}
	return img, nil
}
