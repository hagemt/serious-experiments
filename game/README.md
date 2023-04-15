# GoLang embedding NextJS

Next has the concept of exporting your front-end site for serving over CDN.

This experiment embeds the export in an executable that serves HTTP.

Separately, there's the whispers of using gomobile in cmd/game

## Show Me

Run `make` and you should see an API for HTTP GET /ping and HTML on:

* http://localhost:3000/ui

Other targets are also provided, e.g. `make image` for building an OCI image.
