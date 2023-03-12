package textsynth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImage(t *testing.T) {
	t.SkipNow() // this test spends credits and writes image(s) to /tmp dir:
	prompt := "an astronaut riding a horse"
	client := newClient(t)
	_, err := client.TextToImage(context.Background(), prompt, &ImagerOptions{
		Seed: 0,
	})
	require.NoError(t, err)
}
