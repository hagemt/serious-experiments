package textsynth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type (
	creditsResponse struct {
		Credits json.Number `json:"credits"`
	}
)

func (api *apiClient) Credits(ctx context.Context) (int64, error) {
	to := fmt.Sprintf("%s/v1/credits", api.base.String())
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, to, nil)
	if err != nil {
		return -1, err
	}
	t, err := doRoundTrip[creditsResponse](api, r)
	if err != nil {
		return -1, err
	}
	return t.Credits.Int64()
}
