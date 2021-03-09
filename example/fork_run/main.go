// +build linux

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

		quoted := false
		cmdArgs := strings.FieldsFunc(string(msg), func(r rune) bool {
			if r == '\'' {
				quoted = !quoted
			}
			return !quoted && r == ' '
		})

		// 第一个是执行程序
		initCmd, err := exec.LookPath(cmdArgs[0])
		if err != nil {
			log.Printf("exec LookPath failed. err: %s\n", err.Error())
			return err
		}

		// 对args判断 bash -c 'echo 1212' ===> arg[2] = "echo 1212"，不能有单引号
		for i, cmdArg := range cmdArgs[1:] {
			if -1 != strings.IndexByte(cmdArg, ' ') {
				cmdArgs[i+1] = strings.Trim(cmdArg, "'")
			}
		}
		//cmdArgs[2] = strings.Trim(cmdArgs[2], "'")
		log.Printf("init cmd: {%s}, cmdArgs: %#v\n", initCmd, cmdArgs)

		// execv，已经替换了执行文件，这是最后一条指令，后面都不会执行
		if err := syscall.Exec(initCmd, cmdArgs, os.Environ()); err != nil {
			log.Fatalf(err.Error())
		}
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

	if !tty {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Foreground: false,
		}
	}

	if err := cmd.Start(); err != nil {
		log.Printf("cmd Start failed. err: %s\n", err.Error())
		return err
	}

	// 将exec的命令用管道发送
	var cmdStrBuilder strings.Builder
	for _, arg := range cmdArgs {
		if -1 != strings.IndexByte(arg, ' ') {
			// 这个string中存在空格，需要作为一个整体参数，用单引号括起来
			cmdStrBuilder.WriteByte('\'')
			cmdStrBuilder.WriteString(arg)
			cmdStrBuilder.WriteByte('\'')
		} else {
			cmdStrBuilder.WriteString(arg)
		}
		cmdStrBuilder.WriteByte(' ')
	}

	childExe := strings.TrimSpace(cmdStrBuilder.String())
	log.Printf("child pid: %d running, child exe cmd: {%s}\n", cmd.Process.Pid, childExe)

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
