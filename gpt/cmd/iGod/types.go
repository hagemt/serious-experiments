package main

import (
	"context"
	"fmt"
)

type (
	// divineOption configures iGod
	divineOption func(*iGod) error

	// edict is that which iGod spake
	edict interface {
		Act() error
		fmt.Stringer
	}

	// speaker will prompt iGod for an edict given some context
	speaker func(ctx context.Context, prompt string) edict

	simpleEdict  string
	complexEdict struct {
		ctx   context.Context // passed into:
		actor func(ctx context.Context) error
		text  func(ctx context.Context) string
	}
)
