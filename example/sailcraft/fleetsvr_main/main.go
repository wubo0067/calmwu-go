package main

import (
	"fmt"
	"os"
	"runtime"
	"sailcraft/fleetsvr_main/fleetsvr"

	"github.com/urfave/cli"
)

var (
	version   = "0.0.1"
	buildtime = ""
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("sailcraft version=%s buildtime=%s\n", version, buildtime)
	}

	app := cli.NewApp()
	app.Name = "FleetSvr"
	app.Usage = "SailCraft FleetSvr Service"
	app.Commands = fleetsvr.FleetSvrCmds

	app.Run(os.Args)

}
