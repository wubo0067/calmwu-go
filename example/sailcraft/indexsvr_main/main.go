/*
 * @Author: calmwu
 * @Date: 2017-09-18 09:59:42
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-09-19 11:42:21
 * @Comment:
 */

// 按名字查询服务

package main

import (
	"fmt"
	"os"
	"runtime"
	"sailcraft/indexsvr_main/root"

	"github.com/urfave/cli"
)

var (
	version   = "0.0.1"
	buildtime = ""
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("sandmonk version=%s buildtime=%s\n", version, buildtime)
	}

	app := cli.NewApp()
	app.Name = "MonkeyKing"
	app.Usage = "SailCraft Query Service"
	app.Commands = root.IndexSvrCmds

	app.Run(os.Args)
}
