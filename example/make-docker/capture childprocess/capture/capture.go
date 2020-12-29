/*
 * @Author: calm.wu
 * @Date: 2019-11-10 16:41:02
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-11-10 17:04:10
 */

// https://www.socketloop.com/tutorials/golang-capture-stdout-of-a-child-process-and-act-according-to-the-result

package main

import (
	"bufio"
	"bytes"
	"os/exec"
	"syscall"

	calm_utils "github.com/wubo0067/calmwu-go/utils"
)

func readChildAndKillChildTimeout() {
	logger := calm_utils.NewSimpleLog(nil)

	cmd := &exec.Cmd{
		Path: "../childprocess/childprocess.exe",
		Args: []string{"childprocess.exe"},
	}

	stdout, _ := cmd.StdoutPipe()
	bc := bufio.NewScanner(stdout)

	cmd.Start()

	childProcessDone := make(chan error)
	go func() {
		childProcessDone <- cmd.Wait()
		logger.Printf("wait child completed\n")
		close(childProcessDone)
	}()

L:
	for {
		select {
		case err, ok := <-childProcessDone:
			if ok {
				if err != nil {
					// 子进程exitcode != 0
					logger.Printf("err: %s\n", err.Error())
					if exiterr, ok := err.(*exec.ExitError); ok {
						logger.Printf("exitCode:%d\n", exiterr.ExitCode())
						// The program has exited with an exit code != 0

						// This works on both Unix and Windows. Although package
						// syscall is generally platform dependent, WaitStatus is
						// defined for both Unix and Windows and in both cases has
						// an ExitStatus() method with the same signature.
						if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
							logger.Printf("Exit Status: %d", status.ExitStatus())
						}
					}
				} else {
					// 正常结束
					logger.Printf("err is nil\n")
				}
				break L
			} else {
				logger.Printf("childProcessDone is closed\n")
				break L
			}
		default:
		}

		if bc.Scan() {
			logger.Printf(bc.Text())
		}
	}

	logger.Println("-------------")
}

func realTimeReadChildOutput1() {
	logger := calm_utils.NewSimpleLog(nil)

	cmd := &exec.Cmd{
		Path: "../childprocess/childprocess.exe",
		Args: []string{"childprocess.exe"},
	}

	stdout, _ := cmd.StdoutPipe()
	bc := bufio.NewScanner(stdout)

	cmd.Start()

	for bc.Scan() {
		logger.Printf(bc.Text())
	}

	cmd.Wait()
}

func realTimeReadChildOutput() {
	logger := calm_utils.NewSimpleLog(nil)

	cmd := &exec.Cmd{
		Path: "../childprocess/childprocess.exe",
		Args: []string{"childprocess.exe"},
	}

	stdout, _ := cmd.StdoutPipe()
	reader := bufio.NewReader(stdout)

	cmd.Start()

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			logger.Printf("read error, err:%s", err.Error())
			break
		}
		logger.Printf(line)
	}

	cmd.Wait()
}

func waitForFullOutput() {
	logger := calm_utils.NewSimpleLog(nil)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := &exec.Cmd{
		Path:   "../childprocess/childprocess.exe",
		Args:   []string{"childprocess.exe"},
		Stdout: stdout,
		Stderr: stderr,
	}

	err := cmd.Run()
	if err != nil {
		logger.Fatalf("childprocess Run failed. err:%s", err.Error())
	}

	logger.Printf("childprocess stdout:[%s]", stdout.String())
	logger.Printf("childprocess stderr:[%s]", stderr.String())
}

func main() {
	//waitForFullOutput()
	//realTimeReadChildOutput1()
	readChildAndKillChildTimeout()
}
