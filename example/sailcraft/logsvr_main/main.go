/*
 * @Author: calmwu
 * @Date: 2017-08-31 11:21:43
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-08-31 14:36:41
 * @Comment:
 */
package main

import (
	"fmt"
	"os"
	"runtime"
	"sailcraft/logsvr_main/logsvr"
	"time"

	"github.com/urfave/cli"
)

var (
	version   = "0.0.1"
	buildtime = ""
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("logsvr version=%s buildtime=%s\n", version, buildtime)
	}

	app := cli.NewApp()
	app.Name = "logsvr"
	app.Usage = "SailCraft Log Service"
	app.Flags = logsvr.LogSvrFlag
	app.Action = logsvr.LogSvrAction

	app.Run(os.Args)

	time.Sleep(time.Second)
}
