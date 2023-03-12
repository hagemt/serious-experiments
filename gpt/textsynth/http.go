package textsynth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

// TODO: make these configurable via the environment variables
const (
	defaultBase       = "https://api.textsynth.com"
	defaultEngineName = "gptj_6B" // TEXTSYNTH_DEFAULT_ENGINE_NAME
	key               = ""        // from .env: TEXTSYNTH_KEY (_BASE_URL)
)

func (api *apiClient) failedRoundTrip(in *http.Response, err error) error {
	var status string
	if in != nil {
		a, _ := httputil.DumpRequestOut(in.Request, true)
		b, _ := httputil.DumpResponse(in, true)
		log.Println(string(a))
		log.Println(string(b))

		var e struct {
			Error string `json:"error"`
			//Status json.Number `json:"status"`
		}
		status = in.Status
		if err = json.NewDecoder(in.Body).Decode(&e); err != nil {
			err = fmt.Errorf("json failure=%w", err)
		} else {
			err = fmt.Errorf("API response=%s", e.Error)
		}
	} else {
		status = "?net failure"
	}
	return fmt.Errorf("%s: %w", status, err)
}

func (api *apiClient) prepareRequest(in *http.Request) *http.Request {
	r := in.Clone(in.Context())
	if !in.URL.IsAbs() {
		a := api.base.ResolveReference(in.URL).String()
		r, _ = http.NewRequestWithContext(in.Context(), in.Method, a, in.Body)
	}
	r.Header.Set("authorization", fmt.Sprintf("Bearer %s", api.key))
	r.Header.Set("user-agent", "textsynth-go/0.1")
	return r
}

func (api *apiClient) newPOST(engineName, urlSuffix string, body any) (*http.Request, error) {
	var in bytes.Buffer
	if err := json.NewEncoder(&in).Encode(body); err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/v1/engines/%s/%s", engineName, urlSuffix)
	r, err := http.NewRequest(http.MethodPost, path, &in)
	if err != nil {
		return nil, err
	}
	r.Header.Set("accept", "application/json")
	r.Header.Set("content-type", "application/json")
	return r, nil
}

func doRoundTrip[T any](api *apiClient, in *http.Request) (*T, error) {
	r := api.prepareRequest(in)
	s, err := api.httpClient.Do(r)
	if err != nil {
		return nil, api.failedRoundTrip(s, err)
	}
	defer s.Body.Close()

	if s.StatusCode != http.StatusOK {
		return nil, api.failedRoundTrip(s, err)
	}
	var t T
	if err = json.NewDecoder(s.Body).Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}
