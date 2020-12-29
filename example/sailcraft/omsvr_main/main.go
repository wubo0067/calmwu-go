/*
 * @Author: calmwu
 * @Date: 2018-05-18 11:15:13
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-18 11:15:54
 * @Comment:
 */

package main

import (
	"fmt"
	"os"
	"runtime"
	"sailcraft/omsvr_main/root"

	"github.com/urfave/cli"
)

var (
	version   = "0.0.2"
	buildtime = ""
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("OMSvr version=%s buildtime=%s\n", version, buildtime)
	}

	app := cli.NewApp()
	app.Name = "OMSvr"
	app.Usage = "SailCraft Operation Manager Service"
	app.Commands = root.OMSvrCmds

	app.Run(os.Args)
}
