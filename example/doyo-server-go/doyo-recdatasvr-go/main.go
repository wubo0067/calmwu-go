/*
 * @Author: calmwu
 * @Date: 2018-10-30 16:55:35
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-01 15:59:43
 */

package main

import (
	"doyo-server-go/doyo-recdatasvr-go/recdatasvr"
	"fmt"
	"os"
	"runtime"

	"github.com/urfave/cli"
)

var (
	version   = "0.0.1"
	buildtime = ""
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("doyo-routersvr version=%s buildtime=%s\n", version, buildtime)
	}

	app := cli.NewApp()
	app.Name = "doyo-recdatasvr"
	app.Usage = "doyo Recommended data server"
	app.Flags = recdatasvr.DoyoRecDataSvrFlags
	app.Action = recdatasvr.Main

	app.Run(os.Args)

	fmt.Println("recdatasvr.Main exit!")
}
