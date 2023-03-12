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
		Tokens(ctx context.Context, engineName string) Tokenizer

		Completions(ctx context.Context, engineName string) TextCompleter
		TextToImage(ctx context.Context, prompt string, advanced *ImagerOptions) (Images, error)

		LogProb(ctx context.Context, engineName string) LogProber
		Stranslator(ctx context.Context, simple *TranslationOptions) SimpleTranslator

		String() string
	}

	apiClient struct {
		base *url.URL
		key  string

		httpClient *http.Client
	}

	Settings struct {
		SLA time.Duration // default: one minute
	}
)

var Defaults = struct {
	EngineName string
	TimeToWait time.Duration
}{
	EngineName: "gptj_6B",
	TimeToWait: time.Minute,
	/*
	   gptj_6B: GPT-J is a language model with 6 billion parameters trained on the Pile (825 GB of text data) published by EleutherAI. Its main language is English but it is also fluent in several other languages. It is also trained on several computer languages.
	   boris_6B: Boris is a fine tuned version of GPT-J for the French language. Use this model is you want the best performance with the French language.
	   fairseq_gpt_13B: Fairseq GPT 13B is an English language model with 13 billion parameters. Its training corpus is less diverse than GPT-J but it has better performance at least on pure English language tasks.
	   gptneox_20B: GPT-NeoX-20B is the largest publically available English language model with 20 billion parameters. It was trained on the same corpus as GPT-J.
	   codegen_6B_mono: CodeGen-6B-mono is a 6 billion parameter model specialized to generate source code. It was mostly trained on Python code.
	   m2m100_1_2B: M2M100 1.2B is a 1.2 billion parameter language model specialized for translation. It supports multilingual translation between 100 languages. See the translate endpoint.
	   stable_diffusion: Stable Diffusion is a 1 billion parameter text to image model trained to generate 512x512 pixel images from English text (sd-v1-4.ckpt checkpoint). See the text_to_image endpoint. There are specific use restrictions associated with this model.
	*/
}

func NewClient(key string, optional *Settings) TextSynthAPI {
	b, _ := url.Parse(defaultBase)
	dt := Defaults.TimeToWait // long
	if optional != nil && optional.SLA >= 0 {
		dt = optional.SLA
	}
	return &apiClient{
		base: b,
		key:  key,

		httpClient: &http.Client{
			Timeout: dt,
		},
	}
}

func (api *apiClient) String() string {
	return fmt.Sprintf("GPT: %s", api.base.String())
}
