package textsynth

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImage(t *testing.T) {
	t.SkipNow() // this test spends credits and writes image(s) to temp dir:
	prompt := "an astronaut riding a horse"
	client := newClient(t)
	o, err := client.TextToImage(context.Background(), prompt, &ImagerOptions{
		ImageCount: 4,
		Seed:       1,
	})
	require.NoError(t, err)
	o.SaveAll(os.TempDir())
}
