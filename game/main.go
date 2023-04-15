package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"os"

	"github.com/hagemt/serious_experiments/game/internal"
)

//go:embed all:out/*
var uiExports embed.FS

var version string

func main() {
	addr := internal.EnvString("HTTP_ADDR", "localhost:3000")
	demo := internal.EnvString("HTTP_DEMO", "") == "simple-ui"
	if demo {
		ui, _ := fs.Sub(uiExports, "out") // ./out/* -> .../ui/*
		internal.ServeDemo(context.Background(), addr, ui.(fs.SubFS))
	} else {
		program := os.Args[0] // TODO: other commands, in proper CLI (-ui)
		log.Println(program, "version:", version)
	}
}
