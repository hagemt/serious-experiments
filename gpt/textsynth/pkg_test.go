package textsynth

import (
	"context"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSynth(t *testing.T) {
	t.SkipNow() // this actually expends credits:
	synth := newClient(t)
	prompt := "Once upon a time, there was"
	tc := synth.Completions(context.Background(), Defaults.EngineName)
	c, err := tc.Complete(prompt, &CompleteOptions{
		MaxTokens: 20,
	})
	require.NoError(t, err)
	assert.Len(t, c, 2)
	for _, s := range c {
		assert.True(t, strings.HasPrefix(s.String(), prompt+" a"))
	}
}

func TestSimpleTranslation(t *testing.T) {
	t.SkipNow() // this actually expends credits:
	client := newClient(t)
	simple := client.Stranslator(context.Background(), &TranslationOptions{
		TargetLang: "es",
	})
	o, err := simple.Translate([]string{"Hello!"})
	require.NoError(t, err)
	assert.Len(t, o, 1)
	log.Println("Hola?", o[0].String())
}

func TestFetchLogProb(t *testing.T) {
	t.SkipNow() // this actually expends credits:
	client := newClient(t)
	prober := client.LogProb(context.Background(), Defaults.EngineName)
	p, err := prober.Ask("Will it ", "blend?")
	require.NoError(t, err)
	assert.NotEqual(t, "", p.String())
	log.Println("did it?", p.String())
}
