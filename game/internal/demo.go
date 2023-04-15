package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func EnvString(name, defaultValue string) string {
	if s, ok := os.LookupEnv(name); ok {
		return s
	}
	return defaultValue
}

func WriteJSON[T any](sc int, w http.ResponseWriter, body T) {
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(sc)
	if err := json.NewEncoder(w).Encode(&body); err != nil {
		panic(err)
	}
}

func awaitTermination(ctx context.Context, c io.Closer) {
	dtx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()
	<-dtx.Done()
	_ = c.Close()
}

func tcpListen(ctx context.Context, addr string, my *http.Server) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	go awaitTermination(ctx, my)
	if err := my.Serve(l); err != http.ErrServerClosed {
		return err
	}
	return nil
}

func ServeDemo(ctx context.Context, addr string, ui fs.SubFS) error {
	msg := StatusMessage(ctx)
	mux := http.DefaultServeMux
	if ui != nil {
		uiHandler := http.FileServer(http.FS(ui))
		mux.Handle("/ui/", http.StripPrefix("/ui", uiHandler))
		msg = fmt.Sprintf("UP and /ui on http://%s", addr)
	}
	log.Println("starting services", msg)
	return tcpListen(ctx, addr, &http.Server{
		BaseContext: func(_ net.Listener) context.Context {
			return ctx // augment w/ desired key-values
		},
		Handler:           httpRoutes(mux, msg),
		ReadHeaderTimeout: time.Second,
	})
}

func StatusMessage(ctx context.Context) string {
	if err := ctx.Err(); err != nil {
		return err.Error()
	}
	return "UP"
}
