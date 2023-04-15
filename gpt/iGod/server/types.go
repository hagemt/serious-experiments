package server

import (
	"context"

	"github.com/hagemt/serious_experiments/gpt/iGod/client"
)

const (
	ServiceDeity         = ctxKeyText("iGod")
	ServiceDivineOptions = "iGodNames"
)

type (
	// Service provides HTTP access to iGod
	Service interface {
		ListenAndServe(ctx context.Context, addr string) error
		AddSpeaker(fn client.SpeakerFunc) Service
		Test(ctx context.Context) ServiceEdict
	}

	ServiceEdict interface {
		client.Edict
	}

	ctxKeyText string
)

func (s ctxKeyText) String() string {
	return string(s)
}
