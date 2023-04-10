package textsynth

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type (
	Images        []image.Image
	ImagerOptions struct {
		Prompt string `json:"prompt"` // required: what to draw (e.g. Funny cats)

		ImageCount    int `json:"image_count,omitempty"`    // [1,4]
		Width         int `json:"width,omitempty"`          // {384,512,640,768}
		Height        int `json:"height,omitempty"`         // same limits, H*W <= 512*768
		Timesteps     int `json:"timesteps,omitempty"`      // e.g. 50
		GuidanceScale int `json:"guidance_scale,omitempty"` // default: 7.5
		Seed          int `json:"seed,omitempty"`           // default: 0
	}

	sdOne struct {
		Base64 string `json:"data"`
	}
	sdOut struct {
		Images []sdOne `json:"images"`
	}
)

func closeQuietly(c io.Closer) {
	_ = c.Close()
}

func (is Images) Dump(how string) error {
	// TODO: check TERM
	switch how {
	case "imgcat":
		for _, i := range is {
			var b bytes.Buffer
			if err := jpeg.Encode(&b, i, nil); err != nil {
				continue
			}

			var sb strings.Builder
			sb.WriteRune('\033')
			sb.WriteRune(']')
			sb.WriteString("1337;File=")
			// name=base64encode(filename);
			sb.WriteString("size=")
			sb.WriteString(strconv.FormatInt(int64(b.Len()), 10))
			sb.WriteString(";inline=1:")
			sb.Write(b.Bytes())
			sb.WriteRune('\a')
			fmt.Println(sb.String())
		}
		return nil
	}

	return fmt.Errorf("unknown mechanism: %s", how)
}

func (is Images) SaveAll(dir string) error {
	errs := make([]error, 0, len(is))
	path := filepath.Clean(dir)
	if err := os.MkdirAll(path, os.ModeTemporary); err != nil {
		return fmt.Errorf("mkdir -p %s # failed: %w", path, err)
	}
	for i, img := range is {
		f, err := os.Create(fmt.Sprintf("%s/%d.jpg", path, i))
		if err != nil {
			errs = append(errs, err)
			continue
		}
		defer closeQuietly(f)
		if err = jpeg.Encode(f, img, nil); err != nil {
			errs = append(errs, err)
			continue
		}
		log.Println("saved image:", f.Name())
	}
	return errors.Join(errs...)
}

func (api *apiClient) TextToImage(ctx context.Context, prompt string, advanced *ImagerOptions) (Images, error) {
	var in ImagerOptions
	if advanced != nil {
		in = *advanced
	}
	in.Prompt = prompt
	a, err := api.newPOST("stable_diffusion", "text_to_image", in)
	if err != nil {
		return nil, err
	}
	b, err := doRoundTrip[sdOut](api, a.Clone(ctx))
	if err != nil {
		return nil, err
	}
	errs := make([]error, 0, len(b.Images))
	images := make(Images, len(b.Images))
	for i, img := range b.Images {
		b, berr := base64.StdEncoding.DecodeString(img.Base64)
		if berr != nil {
			errs = append(errs, berr)
			continue
		}
		d, derr := jpeg.Decode(bytes.NewReader(b))
		if derr != nil {
			errs = append(errs, derr)
			continue
		}
		images[i] = d
	}
	//images.SaveAll(os.TempDir())
	return images, errors.Join(errs...)
}
