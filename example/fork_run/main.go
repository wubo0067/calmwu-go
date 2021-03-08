/*
 * @Author: CALM.WU
 * @Date: 2021-03-07 21:40:47
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2021-03-07 21:57:51
 */

package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const usage = "forkRun run cmd"

var runCmd = cli.Command{
	Name:  "run",
	Usage: `Create a child process to exec cmd`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
	}, // 子命令选项不算args
	Action: func(context *cli.Context) error {
		// 这个是子命令的参数
		log.Printf("context Args: %#v\n", context.Args())

		cmdArgs := make([]string, len(context.Args()))
		for i, arg := range context.Args() {
			cmdArgs[i] = arg
		}

		ForkExec(cmdArgs, context.Bool("ti"))
		return nil
	},
}

var initCmd = cli.Command{
	Name:  "init",
	Usage: "Init child process",
	Action: func(context *cli.Context) error {
		// 读取执行的命令
		pipe := os.NewFile(uintptr(3), "pipe")
		defer pipe.Close()
		msg, err := ioutil.ReadAll(pipe)
		if err != nil {
			log.Printf("init read pipe failed. err:%s\n", err.Error())
			return err
		}

		if len(msg) == 0 {
			err = errors.New("read empty from pipe")
			log.Printf("%s\n", err.Error())
			return err
		}

		cmdArgs := strings.Split(string(msg), " ")
		log.Printf("cmdArgs: %#v\n", cmdArgs)

		// 第一个是执行程序
		initCmd, err := exec.LookPath(cmdArgs[0])
		if err != nil {
			log.Printf("exec LookPath failed. err: %s\n", err.Error())
			return err
		}

		log.Printf("find init cmd: %s\n", initCmd)
		// execv
		if err := syscall.Exec(initCmd, cmdArgs[0:], []string{}); err != nil {
			log.Fatalf(err.Error())
		}
		log.Printf("--------init exit--------\n")
		return nil
	},
}

// ForkExec make a duplicate of the current process, execv replaces the
// duplicate parent process with a new process
func ForkExec(cmdArgs []string, tty bool) error {
	log.Printf("cmdArgs: %#v, tty: %v\n", cmdArgs, tty)

	// 子进程命令读取管道
	rPipe, wPipe, err := os.Pipe()
	if err != nil {
		log.Printf("make Pipe failed. err:%s", err.Error())
		return err
	}

	// 自身的执行命令
	selfExe, _ := os.Readlink("/proc/self/exe")
	log.Printf("selfCmd: %s\n", selfExe)

	cmd := &exec.Cmd{
		Path: selfExe,
		Args: []string{selfExe, "init"},
		Dir: func() string {
			cwd, _ := os.Getwd()
			return cwd
		}(),
		Stdin:      os.Stdin,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
		ExtraFiles: []*os.File{rPipe},
	}

	if err := cmd.Start(); err != nil {
		log.Printf("cmd Start failed. err: %s\n", err.Error())
		return err
	}

	log.Printf("child pid: %d running\n", cmd.Process.Pid)

	// 将exec的命令用管道发送
	childExe := strings.Join(cmdArgs, " ")
	wPipe.WriteString(childExe)
	// 关闭管道
	wPipe.Close()
	if tty {
		// 等待子进程退出
		log.Printf("tty for wait child\n")
		cmd.Wait()
		log.Printf("pid:%d child process exit!\n", cmd.Process.Pid)
	} else {
		// os.Stdout.Sync()
		// os.Stderr.Sync()
		cmd.Process.Release()
	}
	log.Printf("--------run exit--------\n")
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "forkRun"
	app.Usage = usage

	app.Commands = []cli.Command{
		runCmd,
		initCmd,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err.Error())
	}
}
