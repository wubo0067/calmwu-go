/*
 * @Author: calmwu
 * @Date: 2018-01-30 16:47:16
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-02-22 19:33:54
 * @Comment:
 */

package main

import (
	"fmt"
	"os"
	"runtime"
	"sailcraft/financesvr_main/root"

	"github.com/urfave/cli"
)

var (
	version   = "0.0.2"
	buildtime = ""
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("FinanceSvr version=%s buildtime=%s\n", version, buildtime)
	}

	app := cli.NewApp()
	app.Name = "FinanceSvr"
	app.Usage = "SailCraft Finance Service"
	app.Commands = root.FinanceSvrCmds

	app.Run(os.Args)
}
