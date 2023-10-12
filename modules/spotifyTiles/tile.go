package spotifyTiles

import (
	"bytes"
	"context"
	"errors"
	"github.com/carlmjohnson/requests"
	"golang.org/x/image/draw"
	"image"
	"image/jpeg"
	"log/slog"
	"math"
	"time"
)

var errTooFewImages = errors.New("too few images")

const outputSize = 300

func generateFromAlbumImages(imgs []*albumImages) (image.Image, error) {
	if len(imgs) <= 3 {
		return nil, errTooFewImages
	}

	numImagesPerSide := 2
	for {
		if x := numImagesPerSide + 1; x*x > len(imgs) {
			break
		} else {
			numImagesPerSide = x
		}
	}

	img := image.NewRGBA(image.Rect(0, 0, outputSize, outputSize))

	subImageDimension := int(math.Ceil(float64(outputSize / numImagesPerSide)))

	slog.Debug("image dimensions calculated", "dim", subImageDimension)

	for x := 0; x < numImagesPerSide; x += 1 {
		for y := 0; y < numImagesPerSide; y += 1 {
			i := (x * numImagesPerSide) + y
			res, err := getImageFromURL(imgs[i].Images[0].Url)
			if err != nil {
				return nil, err
			}
			resized := image.NewRGBA(image.Rect(0, 0, subImageDimension, subImageDimension))
			draw.ApproxBiLinear.Scale(resized, resized.Rect, res, res.Bounds(), draw.Over, nil)
			pasteImage(resized, img, x*subImageDimension, y*subImageDimension)
		}
	}

	return img, nil
	//return nil
}

func getImageFromURL(url string) (image.Image, error) {
	slog.Debug("get image URL", "url", url)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
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

func pasteImage(src image.Image, dst *image.RGBA, atX, atY int) {
	srcBounds := src.Bounds()
	for x := 0; x < srcBounds.Dx(); x += 1 {
		for y := 0; y < srcBounds.Dy(); y += 1 {
			dst.Set(x+atX, y+atY, src.At(x, y))
		}
	}
}
