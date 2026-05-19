// Command gen builds the go-srvc.com static site into a dist directory.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-srvc/website/internal/render"
)

func main() {
	out := flag.String("out", "dist", "output directory")
	serve := flag.Bool("serve", false, "after building, serve dist over http")
	addr := flag.String("addr", ":8080", "serve address")
	flag.Parse()

	if err := render.Build(*out); err != nil {
		slog.Error("build failed", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("build complete", slog.String("out", *out))

	if !*serve {
		return
	}
	slog.Info("serving", slog.String("addr", *addr), slog.String("url", fmt.Sprintf("http://localhost%s", *addr)))
	if err := http.ListenAndServe(*addr, http.FileServer(http.Dir(*out))); err != nil {
		slog.Error("serve failed", slog.Any("error", err))
		os.Exit(1)
	}
}
