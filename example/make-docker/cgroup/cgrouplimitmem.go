// +build linux
/*
 * @Author: calmwu
 * @Date: 2019-03-30 22:08:04
 * @Last Modified by: calmwu
 * @Last Modified time: 2019-03-31 00:20:42
 */

package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strconv"
	"syscall"
	"time"
)

// 挂载了memory subsystem的hierarchy的根目录位置
const cgroupMemoryHierarchyMount = "/sys/fs/cgroup/memory"

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)

	if os.Args[0] == "/proc/self/exe" {
		// 容器进程
		logger.Printf("in container pid %d\n", syscall.Getpid())
		// 定时退出
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		containerSignalChan := make(chan os.Signal, 1)
		signal.Notify(containerSignalChan, syscall.SIGTERM, syscall.SIGINT)
		go func() {
			signal := <-containerSignalChan
			logger.Printf("in container pid:%d receive signal:%v\n", syscall.Getpid(), signal)
			cancel()
		}()

		cmd := exec.CommandContext(ctx, "sh", "-c", "stress --vm-bytes 200m --vm-keep -m 1")
		cmd.SysProcAttr = &syscall.SysProcAttr{}

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			logger.Println("in container", err.Error())
			os.Exit(-1)
		}
		logger.Printf("in container exit!\n")
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)

	cmd := exec.Command("/proc/self/exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		logger.Println("Error:", err.Error())
		os.Exit(-1)
	} else {
		pid := cmd.Process.Pid
		// 得到容器进程在外部的进程id
		logger.Printf("container int host pid[%d]\n", pid)

		go func() {
			signal := <-signalChan
			logger.Printf("pid:%d receive signal:%v\n", syscall.Getpid(), signal)
			// forward signal
			cmd.Process.Signal(signal)
		}()

		// 挂载
		os.Mkdir(path.Join(cgroupMemoryHierarchyMount, "testStressMemLimit"), 0755)
		// 将容器进程加入cgroup中
		ioutil.WriteFile(path.Join(cgroupMemoryHierarchyMount, "testStressMemLimit", "tasks"),
			[]byte(strconv.Itoa(cmd.Process.Pid)), 0644)
		// 限制进程的内存使用
		ioutil.WriteFile(path.Join(cgroupMemoryHierarchyMount, "testStressMemLimit", "memory.limit_bytes"),
			[]byte("100m"), 0644)

		procState, _ := cmd.Process.Wait()
		logger.Printf("procState[%#v] exit\n", procState)
		os.Remove(path.Join(cgroupMemoryHierarchyMount, "testStressMemLimit"))
	}
}
