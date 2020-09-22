/*
 * @Author: calmwu
 * @Date: 2019-06-23 11:18:36
 * @Last Modified by: calm.wu
 * @Last Modified time: 2020-09-22 15:34:20
 */

package utils

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

// https://blog.kowalczyk.info/article/wOYk/advanced-command-execution-in-go-with-osexec.html

// CmdExec 执行命令
func CmdExec(args ...string) (outStr string, errStr string, err error) {
	baseCmd := args[0]
	cmdArgs := args[1:]

	outStr = ""
	errStr = ""

	Debugf("Exec: %v", args)

	var outb, errb bytes.Buffer
	cmd := exec.Command(baseCmd, cmdArgs...)
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // 使子进程拥有自己的 pgid，等同于子进程的 pid，这样子进程就不会继承父进程的进程组
	}

	err = cmd.Run()
	if err != nil {
		ZLog.Errorf("Exec: %v failed! reason: %s", args, err.Error())
		return
	}

	outStr = outb.String()
	errStr = errb.String()
	return
}

// CmdExecCaptureAndShow 捕获输出
func CmdExecCaptureAndShow(args ...string) (outStr string, errStr string, err error) {
	baseCmd := args[0]
	cmdArgs := args[1:]

	Debugf("Exec: %v", args)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd := exec.Command(baseCmd, cmdArgs...)
	// 标准输出
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	var errStdout, errStderr error
	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)

	err = cmd.Start()
	if err != nil {
		Errorf("cmd.Start: %v failed! reason:%s", args, err.Error())
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// 启动routine做标准输出的拷贝
	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
		wg.Done()
	}()

	_, errStderr = io.Copy(stderr, stderrIn)
	wg.Wait()

	err = cmd.Wait()
	if err != nil {
		Errorf("cmd.Run: %v failed! reason:%s", args, err.Error())
		return
	}

	if errStdout != nil || errStderr != nil {
		Errorf("failed to capture stdout and stderr!")
		return
	}

	outStr = stdoutBuf.String()
	errStr = stderrBuf.String()

	return
}

// RunCommand 运行命令
func RunCommand(name string, args ...string) (stdout string, stderr string, exitCode int) {
	Debugf("run command:", name, args)
	var outbuf, errbuf bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err := cmd.Run()
	stdout = outbuf.String()
	stderr = errbuf.String()

	if err != nil {
		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			// This will happen (in OSX) if `name` is not available in $PATH,
			// in this situation, exit code could not be get, and stderr will be
			// empty string very likely, so we use the default fail code, and format err
			// to string and set to stderr
			Debugf("Could not get exit code for failed program: %v, %v", name, args)
			exitCode = -1
			if stderr == "" {
				stderr = err.Error()
			}
		}
	} else {
		// success, exitCode should be 0 if go is ok
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}
	Debugf("command result, stdout: %v, stderr: %v, exitCode: %v", stdout, stderr, exitCode)
	return
}
