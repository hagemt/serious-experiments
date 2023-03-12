package textsynth

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
)

type (
	ImagerOptions struct {
		Prompt string `json:"prompt"`

		ImageCount    int `json:"image_count,omitempty"`    // [1,4]
		Width         int `json:"width,omitempty"`          // {384,512,640,768}
		Height        int `json:"height,omitempty"`         // same limits, H*W <= 512*768
		Timesteps     int `json:"timesteps,omitempty"`      // e.g. 50
		GuidanceScale int `json:"guidance_scale,omitempty"` // default: 7.5
		Seed          int `json:"seed,omitempty"`           // default: 0
	}
	Images []image.Image

	sdOne struct {
		Base64 string `json:"data"`
	}
	sdOut struct {
		Images []sdOne `json:"images"`
	}
)

func (is Images) Save(dir string) error {
	path := filepath.Clean(dir)
	_ = os.MkdirAll(path, os.ModeTemporary)
	for i, img := range is {
		f, err := os.Create(fmt.Sprintf("%s/%d.jpg", path, i))
		if err != nil {
			continue // TODO: panic/return?
		}
		defer f.Close()
		if err = jpeg.Encode(f, img, nil); err != nil {
			continue
		}
		log.Println("saved image:", f.Name())
	}
	return nil
}

func (api *apiClient) TextToImage(ctx context.Context, prompt string, advanced *ImagerOptions) (Images, error) {
	a, err := api.newPOST("stable_diffusion", "text_to_image", &ImagerOptions{
		Prompt: prompt,
	})
	if err != nil {
		return nil, err
	}
	b, err := doRoundTrip[sdOut](api, a.Clone(ctx))
	if err != nil {
		return nil, err
	}
	images := make(Images, len(b.Images))
	for i, img := range b.Images {
		b, berr := base64.StdEncoding.DecodeString(img.Base64)
		if berr != nil {
			continue
		}
		j, jerr := jpeg.Decode(bytes.NewReader(b))
		if jerr != nil {
			continue
		}
		images[i] = j
	}
	images.Save(os.TempDir())
	return images, nil
}
