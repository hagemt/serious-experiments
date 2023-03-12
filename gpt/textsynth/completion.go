package textsynth

import (
	"context"
	"encoding/json"
)

var sane = &CompleteOptions{
	MaxTokens:           100, // about 400 characters for English
	NumberOfCompletions: 2,

	Temperature: 0.9, // API default is 1
	TopK:        40,
	TopP:        0.9,
}

type (
	CompleteOptions struct {
		Input     string `json:"prompt"`
		MaxTokens int    `json:"max_tokens,omitempty"`
		//Stream bool `json:"stream"` // not supported

		Stop                []string `json:"stop,omitempty"`
		NumberOfCompletions int      `json:"n"` // always 2
		Temperature         float64  `json:"temperature,omitempty"`
		TopK                int      `json:"top_k,omitempty"`
		TopP                float32  `json:"top_p,omitempty"`
		// ^ these are "less advanced" options

		LogitBias         map[string]int `json:"logit_bias,omitempty"`         // index => [-100,100]
		PresencePenalty   float32        `json:"presence_penalty,omitempty"`   // [-2,2]
		FrequencyPenalty  float32        `json:"frequency_penalty,omitempty"`  // [-2,2]
		RepetitionPenalty float32        `json:"repetition_penalty,omitempty"` // [1+]; default: 1
		TypicalP          float32        `json:"typical_p,omitempty"`          // [0,1]; default: 1
	}

	completionsResponseBody struct {
		Output []string `json:"text"`
		//ReachedEnd bool `json:"reached_end"` // may be false if "stream" becomes supported

		TruncatedPrompt bool        `json:"truncated_prompt"`
		InputTokens     json.Number `json:"input_tokens"`
		OutputTokens    json.Number `json:"output_tokens"`
	}

	TextCompleter interface {
		Complete(text string, opts *CompleteOptions) ([]Completed, error)
	}

	Completed struct {
		Input, Output string
	}

	completer func(text string, opts *CompleteOptions) ([]Completed, error)
)

func (c *Completed) String() string {
	return c.Input + c.Output
}

func (fn completer) Complete(text string, opts *CompleteOptions) ([]Completed, error) {
	return fn(text, opts)
}

func (api *apiClient) compBody(s string, in *CompleteOptions) any {
	var opts CompleteOptions
	if in != nil {
		opts = *in
	}
	opts.Input = s
	//opts.MaxTokens = sane.MaxTokens
	opts.NumberOfCompletions = sane.NumberOfCompletions
	// TODO: use other sane defaults
	return &opts
}

func (api *apiClient) Completions(ctx context.Context, engineName string) TextCompleter {
	return completer(func(text string, opts *CompleteOptions) ([]Completed, error) {
		a, err := api.newPOST(engineName, "completions", api.compBody(text, opts))
		if err != nil {
			return nil, err
		}
		b, err := doRoundTrip[completionsResponseBody](api, a.Clone(ctx))
		if err != nil {
			return nil, err
		}
		out := make([]Completed, len(b.Output))
		for i, in := range b.Output {
			ref := &out[i]
			ref.Input = text
			ref.Output = in
		}
		return out, nil
	})
}
