package mango

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const api = "https://api.kanye.rest" // GET => {quote: str}
var _ Cache[string, quoted] = &loadingLRU[string, quoted]{}

type quoted struct {
	KanyeSaid string `json:"quote"`
}

func TestToleranceForFailure(t *testing.T) {
	// failure is detectable
	in := "test"
	ctx := context.Background()
	nothing, err := failed[string](errors.New(in)).Await(ctx)
	require.Error(t, err)
	assert.Nil(t, nothing)
	something, e := loaded(in).Await(ctx)
	require.NoError(t, e)
	assert.Equal(t, in, *something)
}

func TestEmission(t *testing.T) {
	quotes := NewTestHarness[quoted](time.Second * 2)
	getQuote := emits(func(ctx context.Context) (*quoted, error) {
		return quotes.Get(ctx, "https://api.kanye.rest")
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	f := getQuote(ctx)
	g, err := f.Await(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, g.KanyeSaid)
}

func TestLoadingCache(t *testing.T) {
	ctx, long := context.Background(), 2*time.Second
	ld := loading(func(in string) (*quoted, error) {
		dtx, cancel := context.WithTimeout(ctx, long)
		defer cancel()
		return httpGet[quoted](dtx, api+"#"+in)
	})

	c := newCIA(CacheOptions{}, ld)
	require.NotNil(t, c)
	u := c.ComputeIfAbsent("1")
	v := c.ComputeIfAbsent("2")
	w := c.ComputeIfAbsent("1")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // FIXME: shouldn't hit timeout...
	vs, err := awaitMap(ctx, map[string]Future[quoted]{
		"first":  u,
		"second": v,
		"third":  w,
	})
	require.NoError(t, err)
	assert.Equal(t, 3, len(vs))
	// second quote may be equal to first (and third) or not, but WLOG:
	assert.Equal(t, vs["first"].KanyeSaid, vs["third"].KanyeSaid)
}

type testTransport func(*http.Request) (*http.Response, error)

func (tt testTransport) RoundTrip(in *http.Request) (*http.Response, error) {
	return tt(in)
}

func httpGet[T any](ctx context.Context, href string) (*T, error) {
	var transport http.RoundTripper
	debug := http.DefaultTransport.(*http.Transport).Clone()
	debug.ForceAttemptHTTP2 = false
	transport = testTransport(func(in *http.Request) (*http.Response, error) {
		log.Println(in.Method, in.URL)
		return debug.RoundTrip(in)
	})
	thing := NewTestHarness[T](time.Second)
	thing.http.Transport = transport
	return thing.Get(ctx, href)
}

type TestHarness[T any] struct {
	http http.Client
}

func NewTestHarness[T any](timeout time.Duration) *TestHarness[T] {
	th := new(TestHarness[T])
	th.http = *http.DefaultClient
	th.http.Timeout = timeout
	return th
}

func (th *TestHarness[T]) Get(ctx context.Context, href string) (*T, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, href, nil)
	if err != nil {
		return nil, fmt.Errorf("bad request %s %s: %w", r.Method, r.URL.String(), err)
	}
	s, err := th.http.Do(r)
	if err != nil {
		return nil, fmt.Errorf("bad round-trip: %w", err)
	}
	defer func(c io.Closer) {
		_ = c.Close()
	}(s.Body)
	if s.StatusCode != http.StatusOK {
		method, uri := s.Request.Method, s.Request.URL.String()
		err := fmt.Errorf("bad %s %s: %s", method, uri, s.Status)
		return nil, err
	}
	var out T
	if err := json.NewDecoder(s.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
