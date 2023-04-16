package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"os"
	"strings"

	"github.com/hagemt/serious_experiments/game/internal"
)

//go:embed all:out/*
var uiFiles embed.FS

var (
	program string
	version string
)

func init() {
	if program == "" {
		program = strings.TrimPrefix(os.Args[0], "./")
	}
	if version == "" {
		version = "source"
	}
}

func main() {
	addr := internal.EnvString("HTTP_ADDR", "127.0.0.1:3000")
	demo := internal.EnvString("HTTP_DEMO", "") == "simple-ui"
	if demo {
		ui, _ := fs.Sub(uiFiles, "out") // demo: ./out/* -> .../ui/*
		internal.ServeDemo(context.Background(), addr, ui.(fs.SubFS))
	} else {
		// TODO: other commands, in proper CLI (-ui)
		log.Println(program, "version:", version)
	}
}
