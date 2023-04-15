package internal

import (
	"log"
	"net/http"
	"strings"
)

type pong struct {
	Message string `json:"message"`
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
			if up != "UP" && strings.HasPrefix(r.RequestURI, "/ui/") {
				return
			}
			WriteJSON(http.StatusNotFound, w, &struct {
				Message string `json:"message"`
				Suggest string `json:"suggestion"`
			}{
				Message: "Not Found",
				Suggest: up,
			})
		}
	})
	mux.Handle("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(http.StatusOK, w, &pong{
			Message: up,
		})
	}))
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
