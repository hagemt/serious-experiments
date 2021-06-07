package client

import (
	"context"
	"fmt"
	"github.com/boynton/repl"
)

type (
	// DivineOption configures iGod
	DivineOption func(*iGod) error

	// Edict is that which iGod spake
	Edict interface {
		Act() error
		fmt.Stringer
	}

	// Speaker is kind of like a router that handles requests into responses, but for a deity that returns an Edict
	Speaker interface {
		Add(SpeakerFunc) Speaker
		Speak(context.Context, string) Edict
		repl.ReplHandler
	}

	// SpeakerFunc will prompt iGod for an Edict given some context
	SpeakerFunc func(ctx context.Context, prompt string) Edict

	// SimpleEdict captures no Act, only String dialog
	SimpleEdict string

	complexEdict struct {
		ctx   context.Context // passed into:
		actor func(ctx context.Context) error
		text  func(ctx context.Context) string
	}
)
