package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// iGod implements repl.ReplHandler
type iGod struct {
	Speak speaker

	history []string
	idDeity string
	idHuman string
}

func newDeity(options ...divineOption) *iGod {
	deity := &iGod{}
	deity.Reset()
	for _, o := range options {
		_ = o(deity)
	}
	return deity
}

func withNames(name, human string) divineOption {
	return func(god *iGod) error {
		god.idDeity = name
		god.idHuman = human
		return nil
	}
}

func (god *iGod) Complete(_ string) (string, []string) {
	return "", nil // no completions
}

func (god *iGod) Reset() {
	god.history = god.Start()
}

func (god *iGod) Prompt() string {
	return fmt.Sprintf("%s: ", god.idHuman)
}

func (god *iGod) Start() []string {
	return make([]string, 0, 100)
}

func (god *iGod) Stop(history []string) {
	god.history = append(god.history, history...)
}

// Eval handles the expression and returns a string result
func (god *iGod) Eval(input string) (string, bool, error) {
	line := strings.TrimSpace(input)
	if len(line) == 0 {
		return "", false, errors.New("what")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()
	if s := god.Speak(ctx, line); s != nil {
		return s.String(), false, s.Act()
	}
	return "", false, errors.New("no")
}