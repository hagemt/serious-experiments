package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"os"

	"github.com/bijection/ios/internal"
)

//go:embed all:out/*
var uiExports embed.FS

var version string

func main() {
	addr := internal.EnvString("HTTP_ADDR", "localhost:3000")
	dist := internal.EnvString("HTTP_DEMO", "") == "simple-ui"
	if dist {
		ui, _ := fs.Sub(uiExports, "out") // ./out/* -> .../ui/*
		internal.Demo(context.Background(), addr, ui.(fs.SubFS))
		return
	}
	log.Println(os.Args[0], "version:", version)
}
