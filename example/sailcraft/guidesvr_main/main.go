package main

import (
	"fmt"
	"os"
	"runtime"
	"sailcraft/guidesvr_main/root"

	"github.com/urfave/cli"
)

var (
	version   = "0.0.2"
	buildtime = ""
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("GuideSvr version=%s buildtime=%s\n", version, buildtime)
	}

	app := cli.NewApp()
	app.Name = "guidesvr"
	app.Usage = "SailCraft Guide Service"
	app.Commands = root.GuideSvrCmds

	app.Run(os.Args)
}
