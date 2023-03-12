package textsynth

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

type (
	LogProb   fmt.Stringer
	LogProber interface {
		Ask(question, answer string) (LogProb, error)
	}

	logprober func(question, answer string) (LogProb, error)
	logprob   struct {
		EngineName string

		InputContext      string `json:"context"`      // empty string allowed (EOT)
		InputContinuation string `json:"continuation"` // non-empty

		ans float64
		get sync.Once
		got *logprobResponseBody
		// TODO: what's a better respresentation?
	}

	logprobResponseBody struct {
		LogProb     json.Number `json:"logprob"`
		NumTokens   json.Number `json:"num_tokens"`
		IsGreedy    bool        `json:"is_greedy"`
		InputTokens json.Number `json:"input_tokens"`
	}
)

func (fn logprober) Ask(question, answer string) (LogProb, error) {
	return fn(question, answer)
}

func (lp *logprob) String() string {
	lp.get.Do(func() {
		if lp.got != nil {
			lp.ans, _ = lp.got.LogProb.Float64()
		}
	})
	return fmt.Sprintf("%f: %s%s", lp.ans, lp.InputContext, lp.InputContinuation)
}

func (api *apiClient) LogProb(ctx context.Context, engineName string) LogProber {
	return logprober(func(question, answer string) (LogProb, error) {
		in := &logprob{
			InputContext:      question,
			InputContinuation: answer,
		}
		a, err := api.newPOST(engineName, "logprob", in)
		if err != nil {
			return nil, err
		}
		b, err := doRoundTrip[logprobResponseBody](api, a.Clone(ctx))
		if err != nil {
			return nil, err
		}
		in.got = b
		//in.EngineName = engineName
		return in, nil
	})
}
