package textsynth

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newClient(t *testing.T) TextSynthAPI {
	err := godotenv.Load("/Users/teh/Code/MINE/serious-experiments/gpt/.env")
	if err != nil {
		t.FailNow()
	}
	if s, ok := os.LookupEnv("TEXTSYNTH_KEY"); ok {
		return NewClient(s, nil)
	}
	t.FailNow()
	return nil
}

func TestCredits(t *testing.T) {
	synth := newClient(t)
	n, err := synth.Credits(context.Background())
	require.NoError(t, err)
	assert.True(t, n > 0)
}

func TestTokenize(t *testing.T) {
	synth := newClient(t)
	api := synth.Indices(context.Background(), Defaults.EngineName)
	out, err := api.Tokenize("The quick brown fox jumps over the lazy dog")
	require.NoError(t, err)

	tokens := map[string]int64{
		"The":   464,
		"quick": 2068,
		"brown": 7586,
		"fox":   21831,
		"jumps": 18045,
		"over":  625,
		"the":   262,
		"lazy":  16931,
		"dog":   3290,
	}
	for word, index := range out.guess {
		msg := fmt.Sprintf("index of token[%s] in %s", word, Defaults.EngineName)
		assert.Equal(t, index, tokens[word], msg)
	}
}
