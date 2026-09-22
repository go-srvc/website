package main

import (
	"context"

	"github.com/go-srvc/srvc"
)

func main() {
	srvc.RunAndExit(
		// Called during init, Run blocks until shutdown.
		srvc.InitMod("migrations", runMigrations),
		// Init and stop pair without any work of its own during run.
		srvc.IdleMod("cache", openCache, closeCache),
		// Called during run with a context that is canceled on shutdown.
		srvc.CtxMod("consumer", consume),
		// Called during run, service shuts down once it returns.
		srvc.RunMod("job", doWork),
		// Called during stop, in reverse order like any other module.
		srvc.StopMod("flush", flushBuffers),
	)
}

func runMigrations() error { return nil }
func openCache() error     { return nil }
func closeCache() error    { return nil }
func doWork() error        { return nil }
func flushBuffers() error  { return nil }

func consume(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
