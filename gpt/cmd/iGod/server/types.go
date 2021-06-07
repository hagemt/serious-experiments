package server

import (
	"context"
	"github.com/hagemt/bijection/gpt/cmd/iGod/client"
)

const (
	ServiceDeity         = "iGod"
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
)
