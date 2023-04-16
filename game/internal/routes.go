package internal

import (
	"crypto/subtle"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type (
	pong struct {
		Message string `json:"message"`
	}

	errorBody struct {
		Message string `json:"message"`
		Suggest string `json:"suggestion,omitempty"`
	}
)

var healthOK prometheus.Counter

func init() {
	goStats := collectors.NewGoCollector()
	process := collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})
	prometheus.DefaultRegisterer.Unregister(goStats)
	prometheus.DefaultRegisterer.Unregister(process)

	prefix := EnvString("METRICS_PREFIX", "mango_")
	defaults := prometheus.DefaultRegisterer // no namespace
	prefixed := prometheus.WrapRegistererWithPrefix(prefix, defaults)
	prometheus.DefaultRegisterer = prefixed

	healthOK = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pinged",
		Help: "status 200 OK on main health check",
	})
}

func authRequiredMetrics(in string) http.HandlerFunc {
	buf := []byte(in) // expected username:password
	metrics := promhttp.Handler()
	basicAuth := func(r *http.Request) bool {
		var auth []byte
		if username, password, ok := r.BasicAuth(); ok {
			auth = []byte(username + ":" + password)
		}
		return subtle.ConstantTimeCompare(auth, buf) == 0
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if basicAuth(r) {
			metrics.ServeHTTP(w, r)
			return
		}
		w.Header().Add("content-type", "text/plain")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("bad/missing Authorization"))
	}
}

func httpRoutes(mux *http.ServeMux, up string) http.Handler {
	// other API routes can be added here, in addition to GET /ping
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/":
			http.Redirect(w, r, "/ping", http.StatusTemporaryRedirect)
		case "/favicon.ico":
			w.WriteHeader(http.StatusOK)
		default:
			WriteJSON(http.StatusNotFound, w, &errorBody{
				Message: "Not Found",
				Suggest: up,
			})
		}
	})
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		healthOK.Inc()
		WriteJSON(http.StatusOK, w, &pong{
			Message: up,
		})
	})

	metricsAuth := EnvString("METRICS_AUTH", "username:password")
	mux.HandleFunc("/v1/metrics", authRequiredMetrics(metricsAuth))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("HTTP", r.Method, r.RequestURI)
		defer func() {
			if err := recover(); err != nil {
				log.Println("HTTP", err)
			}
		}()
		mux.ServeHTTP(w, r)
	})
}
