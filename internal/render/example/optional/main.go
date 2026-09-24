package main

import (
	"net/http"
	"os"

	"github.com/go-srvc/mods/httpmod"
	"github.com/go-srvc/mods/metermod"
	"github.com/go-srvc/mods/sigmod"
	"github.com/go-srvc/srvc"
)

func main() {
	srvc.RunAndExit(
		sigmod.New(),
		// Dropped before init when METRICS is unset, logged as disabled.
		srvc.Optional(os.Getenv("METRICS") == "1", metermod.New()),
		srvc.Optional(os.Getenv("DEBUG") == "1", debugServer()),
		httpmod.New(httpmod.WithAddr(":8080")),
	)
}

func debugServer() srvc.Module {
	return httpmod.New(
		httpmod.WithAddr("localhost:6060"),
		httpmod.WithHandler(http.DefaultServeMux),
	)
}
