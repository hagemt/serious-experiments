package textsynth

import (
	"context"
	"fmt"
)

type (
	TranslationOptions struct {
		Input []string `json:"text"`

		SourceLang string `json:"source_lang"` // e.g. "auto" or "en" etc.
		TargetLang string `json:"target_lang"` // required two-letter code

		//NumBeams int `json:"num_beams"` // default: 4 (1-5)
		//SplitSentences bool `json:"split_sentences"` // default: true

		//EngineName string // m2m100_1_2B by default
	}

	tlOut struct {
		Text string `json:"text"`
		Lang string `json:"detected_source_lang"`

		//InputTokens  int `json:"input_tokens"`
		//OutputTokens int `json:"output_tokens"`
	}
	tlRes struct {
		Translations []tlOut `json:"translations"`
	}

	Translated       []fmt.Stringer
	SimpleTranslator interface {
		Translate(batch []string) (Translated, error)
	}

	simple string
	stFunc func(batch []string) (Translated, error)
)

func (fn stFunc) Translate(batch []string) (Translated, error) {
	return fn(batch)
}

func (s simple) String() string {
	return string(s)
}

func (api *apiClient) SimpleT(ctx context.Context, opts *TranslationOptions) SimpleTranslator {
	sourceLang := "auto"
	var targetLang string
	if opts != nil && opts.SourceLang != "" {
		sourceLang = opts.SourceLang
		targetLang = opts.TargetLang
	} else if opts != nil {
		targetLang = opts.TargetLang
	} else {
		targetLang = "en"
	}
	return stFunc(func(batch []string) (Translated, error) {
		a, err := api.newPOST("m2m100_1_2B", "translate", &TranslationOptions{
			Input: batch,

			SourceLang: sourceLang,
			TargetLang: targetLang,
		})
		if err != nil {
			return nil, err
		}
		b, err := doRoundTrip[tlRes](api, a.Clone(ctx))
		if err != nil {
			return nil, err
		}
		out := make(Translated, len(b.Translations))
		for i, ts := range b.Translations {
			out[i] = simple(fmt.Sprintf("%s [%s => %s] %s", batch[i], ts.Lang, targetLang, ts.Text))
		}
		return out, nil
	})
}
