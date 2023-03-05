package client

import (
	"context"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
)

/*
func boundedEdict(text string, dt time.Duration, act func (context.Context) error) Edict {
	return &complexEdict{
		actor: func(ctx context.Context) error {
			stop := make(chan error, 1)
			dtx, cancel := context.WithTimeout(ctx, dt)
			go func() {
				stop <- act(dtx)
				cancel()
			}()
			select {
			case err := <-stop: return err
			case <-dtx.Done(): return nil
			}
		},
		text: func(ctx context.Context) string {
			return text
		},
	}
}
*/

func (e customEdict) Act(ctx context.Context) error {
	return e.actor(ctx)
}

func (e customEdict) String() string {
	return e.text(context.TODO())
}

func PromptFromNowhere(ctx context.Context, no func(context.Context) fmt.Stringer) Edict {
	out := &customEdict{
		actor: func(ctx context.Context) error {
			return fmt.Errorf("action failed: %w", ctx.Err())
		},
		text: func(ctx context.Context) string {
			if err := ctx.Err(); err != nil {
				return fmt.Sprintf("What do you think about this? %s", err.Error())
			} else {
				return no(ctx).String()
			}
		},
	}
	return out
}

func FailedEdict(e error) Edict {
	return &customEdict{
		actor: func(_ context.Context) error {
			return e
		},
		text: func(_ context.Context) string {
			return e.Error()
		},
	}
}

func (str SimpleEdict) Act(ctx context.Context) error {
	return ctx.Err()
}

func (str SimpleEdict) String() string {
	return string(str)
}

func Confirm(other Edict) Edict {
	return &customEdict{
		actor: func(ctx context.Context) error {
			var ok bool
			text := &survey.Confirm{
				Message: "Are you sure?",
			}
			if err := survey.AskOne(text, &ok); err != nil {
				return err
			}
			return other.Act(ctx)
		},
		text: func(ctx context.Context) string {
			return other.String()
		},
	}
}
