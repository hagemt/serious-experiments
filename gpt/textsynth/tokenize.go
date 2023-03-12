package textsynth

import (
	"context"
	"encoding/json"
	"strings"
)

type (
	tokenizeResponseBody struct {
		Tokens []json.Number `json:"tokens"`
	}

	Tokenized struct {
		Input string `json:"text"`

		guess map[string]int64
	}

	Tokenizer interface {
		Tokenize(text string) (*Tokenized, error)
	}

	tokenizerFunc func(text string) (*Tokenized, error)
)

func (t *Tokenized) Get(s string) (int64, bool) {
	i, ok := t.guess[s]
	return i, ok
}

func (fn tokenizerFunc) Tokenize(text string) (*Tokenized, error) {
	return fn(text)
}

func (api *apiClient) Indices(ctx context.Context, engineName string) Tokenizer {
	return tokenizerFunc(func(text string) (*Tokenized, error) {
		body := &Tokenized{Input: text}
		a, err := api.newPOST(engineName, "tokenize", body)
		if err != nil {
			return nil, err
		}
		b, err := doRoundTrip[tokenizeResponseBody](api, a.Clone(ctx))
		if err != nil {
			return nil, err
		}
		s := strings.Split(text, " ")
		// assert len(s) == len(b.Tokens)
		tokens := make(map[string]int64, len(s))
		for i, t := range b.Tokens {
			u, _ := t.Int64()
			tokens[s[i]] = u
		}
		body.guess = tokens
		return body, nil
	})
}
