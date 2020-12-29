/*
 * @Author: calmwu
 * @Date: 2018-01-10 15:44:03
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-01-11 16:16:36
 * @Comment:
 */

package main

import (
	"fmt"
	"os"
	"runtime"
	"sailcraft/csssvr_main/root"

	"github.com/urfave/cli"
)

var (
	version   = "0.0.2"
	buildtime = ""
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("CassandraSvr version=%s buildtime=%s\n", version, buildtime)
	}

	app := cli.NewApp()
	app.Name = "cassandra proxy svr"
	app.Usage = "SailCraft Cassandra Service"
	app.Commands = root.CassandraSvrCmds

	app.Run(os.Args)
}
