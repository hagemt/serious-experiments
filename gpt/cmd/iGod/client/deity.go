package client

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// iGod implements Speaker
type iGod struct {
	speaker map[string][]SpeakerFunc
	history []string
	idDeity string
	idHuman string
}

func NewDeity(options ...DivineOption) Speaker {
	deity := &iGod{
		speaker: make(map[string][]SpeakerFunc, 1),
	}
	deity.Reset()
	for _, o := range options {
		_ = o(deity)
	}
	return deity
}

func WithNames(name, human string) DivineOption {
	return func(god *iGod) error {
		god.idDeity = name
		god.idHuman = human
		return nil
	}
}

func (god *iGod) Add(s SpeakerFunc) Speaker {
	key := "" // default prefix = none
	ss, ok := god.speaker[key]
	if !ok {
		ss = make([]SpeakerFunc, 0, 1)
	}
	god.speaker[key] = append(ss, s)
	return god
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

func (god *iGod) Speak(ctx context.Context, line string) Edict {
	for _, ss := range god.speaker {
		for _, s := range ss {
			return s(ctx, line)
		}
	}
	return SimpleEdict(fmt.Sprintf("%s: ...", god.idDeity))
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
		return "", false, errors.New("try: What is your name? (etc.)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	if s := god.Speak(ctx, line); s != nil {
		// TODO: allow control commands
		return s.String(), false, s.Act()
	}
	return "", false, errors.New("no")
}
