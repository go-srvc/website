// Command gen builds the go-srvc.com static site into a dist directory.
package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/go-srvc/website/internal/render"
)

func main() {
	out := flag.String("out", "dist", "output directory")
	cache := flag.String("cache", ".cache", "cache directory for cloned source; empty disables doc rendering")
	flag.Parse()

	if err := render.Build(render.Options{Out: *out, Cache: *cache}); err != nil {
		slog.Error("build failed", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("build complete", slog.String("out", *out))
}
