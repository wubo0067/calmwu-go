package fleetsvr

import (
	"os"
	"os/signal"
	"sailcraft/base"
	"syscall"
)

func waitForSigUsr1() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGUSR1)
	go func() {
		for range sc {
			base.DumpStacks()
		}
	}()
}
