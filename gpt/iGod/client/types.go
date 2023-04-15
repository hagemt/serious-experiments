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
	// (or that which it has done or will do)
	// each "captures" some action (which can be as simple as responding/a delayed response)
	Edict interface {
		Act(context.Context) error
		fmt.Stringer
	}

	// each Speaker routes context + input into Edicts
	// it could be a "human" or other entity (e.g. a diety)
	Speaker interface {
		Add(SpeakerFunc) Speaker
		Speak(ctx context.Context, input string) Edict
		repl.ReplHandler
	}

	// SpeakerFunc will prompt iGod for an Edict given some context
	SpeakerFunc func(ctx context.Context, prompt string) Edict

	// SimpleEdict captures no Act, only String dialog
	SimpleEdict string

	customEdict struct {
		actor func(ctx context.Context) error
		text  func(ctx context.Context) string
	}
)
