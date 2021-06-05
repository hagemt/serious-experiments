package main

import (
	"context"
)

/*
func boundedEdict(text string, dt time.Duration, act func (context.Context) error) edict {
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

func failedEdict(e error) edict {
	return &complexEdict{
		actor: func(ctx context.Context) error {
			return e
		},
		text: func(ctx context.Context) string {
			return e.Error()
		},
	}
}

func (e complexEdict) Act() error {
	return e.actor(e.ctx)
}

func (e complexEdict) String() string {
	return e.text(e.ctx)
}

func (str simpleEdict) Act() error {
	return nil
}

func (str simpleEdict) String() string {
	return string(str)
}
