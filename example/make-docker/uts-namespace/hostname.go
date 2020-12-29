/*
 * @Author: calmwu
 * @Date: 2019-03-16 00:33:30
 * @Last Modified by: calmwu
 * @Last Modified time: 2019-03-16 01:45:29
 */

package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	cmd := exec.Command("sh")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS,
	}

	log.Printf("pid: %d\n", os.Getpid())

	// 设置环境变量
	cmd.Env = []string{"PS1=-[ns-process]- # "}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 在父进程中 读取自己的utsnamespace
	utsNamespace, err := os.Readlink("/proc/self/ns/uts")
	if err != nil {
		log.Fatal("uts readlinke failed! ", err.Error())
	}

	// 获取执行程序路径
	path, err := os.Readlink("/proc/self/exe")
	if err != nil {
		log.Fatal("self exe readlinke failed! ", err.Error())
	}

	log.Printf("%s, %s\n", utsNamespace, path)

	// 这里可以启动多个
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	log.Printf("pid: %d after start\n", os.Getpid())

	cmd.Wait()
	log.Printf("pid: %d\n", os.Getpid())
}
