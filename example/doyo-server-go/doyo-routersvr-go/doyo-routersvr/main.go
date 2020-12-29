/*
 * @Author: calmwu
 * @Date: 2018-09-20 11:28:35
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-09-28 20:09:43
 */

package main

import (
	"fmt"
	"os"
	"runtime"

	"doyo-server-go/doyo-routersvr-go/doyo-routersvr/routersvr"

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
	app.Name = "doyo-routersvr"
	app.Usage = "doyo kafka router server"
	app.Flags = routersvr.RouterSvrFlags
	app.Action = routersvr.Main

	app.Run(os.Args)
	fmt.Println("routersvr.Main exit!")
}
