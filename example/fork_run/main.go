/*
 * @Author: CALM.WU
 * @Date: 2021-03-07 21:40:47
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2021-03-07 21:57:51
 */

package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
)

const usage = "forkRun run cmd"

var runCmd = cli.Command{
	Name:  "run",
	Usage: `Create a child process to run cmd`,
	Action: func(context *cli.Context) error {
		// 这个是子命令的参数
		log.Printf("context Args: %#v\n", context.Args())

		cmdArgs := make([]string, len(context.Args()))
		for i, arg := range context.Args() {
			cmdArgs[i] = arg
		}

		runChild(cmdArgs)
		return nil
	},
}

func runChild(cmdArgs []string) {
	log.Printf("cmdArgs: %#v\n", cmdArgs)
}

func main() {
	app := cli.NewApp()
	app.Name = "forkRun"
	app.Usage = usage

	app.Commands = []cli.Command{
		runCmd,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err.Error())
	}
}
