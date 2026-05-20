package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/XSAM/otelsql"
	"github.com/go-srvc/mods/httpmod"
	"github.com/go-srvc/mods/logmod"
	"github.com/go-srvc/mods/sigmod"
	"github.com/go-srvc/mods/sqlxmod"
	"github.com/go-srvc/mods/tracemod"
	"github.com/go-srvc/srvc"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	db := sqlxmod.New(
		sqlxmod.WithOtel("pgx", os.Getenv("DSN"),
			otelsql.WithAttributes(semconv.DBSystemNamePostgreSQL),
		),
	)
	srvc.RunAndExit(
		logmod.New(),
		tracemod.New(),
		sigmod.New(os.Interrupt),
		db,
		httpmod.New(
			httpmod.WithAddr(":8080"),
			httpmod.WithHandler(handler(db)),
		),
	)
}

func handler(db *sqlxmod.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var version string
		if err := db.DB().GetContext(r.Context(), &version, "SELECT version()"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "db version: %s\n", version)
	})
}
