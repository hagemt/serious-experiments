package textsynth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type (
	TextSynthAPI interface {
		Credits(ctx context.Context) (int64, error)
		Indices(ctx context.Context, engineName string) Tokenizer

		Completions(ctx context.Context, engineName string) TextCompleter
		TextToImage(ctx context.Context, prompt string, advanced *ImagerOptions) (Images, error)

		LogProb(ctx context.Context, engineName string) LogProber
		SimpleT(ctx context.Context, simple *TranslationOptions) SimpleTranslator

		OK(ctx context.Context) (int64, error)
		String() string
	}

	apiClient struct {
		authHeader string
		userAgent  string

		base *url.URL
		http *http.Client
	}

	// Settings allows for overriding Defaults
	Settings struct {
		SLA time.Duration

		BaseURL   string
		UserAgent string
	}
)

// NewClient interacts with TextSynth via the API keyRequired and optional Settings
func NewClient(keyRequired string, optional *Settings) TextSynthAPI {
	if keyRequired == "" {
		panic(fmt.Errorf("ignored Settings: %+v (missing: TextSynth API key)", optional))
	}
	base := Defaults.BaseURL
	if optional != nil && optional.BaseURL != "" {
		base = optional.BaseURL
	}
	b, err := url.Parse(base)
	if err != nil {
		panic(err)
	}
	dt := Defaults.MaxWaitTime // too long?
	if optional != nil && optional.SLA >= 0 {
		dt = optional.SLA
	}
	you := Defaults.UserAgent
	if optional != nil && optional.UserAgent != "" {
		you = optional.UserAgent
	}
	return &apiClient{
		authHeader: fmt.Sprintf("Bearer %s", keyRequired),
		base:       b,
		userAgent:  you,
		http: &http.Client{
			Timeout: dt,
		},
	}
}

func (api *apiClient) OK(ctx context.Context) (int64, error) {
	lo := int64(1000 * 1000) // currently: one USD
	c, err := api.Credits(ctx)
	if err != nil || c <= lo {
		return c, fmt.Errorf("%s low credits, or: %w", api, err)
	}
	return c, nil
}

func (api *apiClient) String() string {
	return fmt.Sprintf("@%s/v1", api.base.String())
}
