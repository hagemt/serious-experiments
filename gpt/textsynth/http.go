package textsynth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
)

const (
	envKey                  = "TEXTSYNTH_KEY"                   // required
	envKeyBase              = "TEXTSYNTH_BASE_URL"              // https://api.textsynth.com
	envKeyDefaultEngineName = "TEXTSYNTH_DEFAULT_ENGINE_NAME"   // vs. gptj_6B
	envKeyUserAgent         = "TEXTSYNTH_SET_CUSTOM_USER_AGENT" // vs. go-textsynth/v...

	envKeyMaxTime = "TEXTSYNTH_SLA"
	envKeyVerbose = "TEXTSYNTH_DEBUG"
)

var (
	envVerbose = envString(envKeyVerbose, "")

	debugBody = strings.HasPrefix(envVerbose, "http+body")
	debugHTTP = strings.HasPrefix(envVerbose, "http")
)

type apiError struct {
	ErrorMessage string `json:"error"`
	//Status json.Number `json:"status"`
}

func (err *apiError) Error() string {
	return err.ErrorMessage
}

func (api *apiClient) failedRoundTrip(in *http.Response, err error) error {
	if in != nil {
		if debugHTTP {
			a, _ := httputil.DumpRequestOut(in.Request, debugBody)
			log.Println("failed", string(a))
			b, _ := httputil.DumpResponse(in, debugBody)
			log.Println("failed", string(b))
		}

		var body apiError
		if err = json.NewDecoder(in.Body).Decode(&body); err != nil {
			return fmt.Errorf("%s: json failure=%w", in.Status, err)
		}
		return fmt.Errorf("%s: API response=%w", in.Status, &body)
	}
	return fmt.Errorf("no round-trip; HTTP failure: %w", err)
}

func (api *apiClient) prepareRequest(in *http.Request) *http.Request {
	in.URL = api.base.ResolveReference(in.URL)
	in.Header.Set("authorization", api.authHeader)
	in.Header.Set("user-agent", api.userAgent)
	return in
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
	s, err := api.http.Do(r)
	if err != nil {
		return nil, api.failedRoundTrip(s, err)
	}
	defer closeQuietly(s.Body)

	if s.StatusCode != http.StatusOK {
		return nil, api.failedRoundTrip(s, err)
	}
	var t T
	if err = json.NewDecoder(s.Body).Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}
