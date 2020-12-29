// +build linux
/*
 * @Author: calmwu
 * @Date: 2019-03-16 17:07:36
 * @Last Modified by: calmwu
 * @Last Modified time: 2019-03-16 18:19:51
 */

package main

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/urfave/cli"
)

var (
	appCmds = []cli.Command{
		{
			Name:   "run",
			Usage:  "start a root container, set uts namespace",
			Flags:  RunFlags,
			Action: startRootContainerAction,
		},
		{
			Name:   "root",
			Usage:  "root container init",
			Flags:  RunFlags,
			Action: rootContainerAction,
		},
		{
			Name:   "child",
			Usage:  "child container init",
			Action: childContainerAction,
		},
	}

	RunFlags = []cli.Flag{
		cli.IntFlag{
			Name:  "count, c",
			Value: 1,
			Usage: "child process count",
		},
	}

	logger = log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)
)

func startRootContainerAction(c *cli.Context) error {
	utsNamespace, _ := os.Readlink("/proc/self/ns/uts")
	hostName, _ := os.Hostname()
	logger.Printf("start uts namespace:%s, hostName:%s\n", utsNamespace, hostName)

	// 启动rootcontainer
	rcArgs := []string{"root", "-c", strconv.Itoa(c.Int("count"))}
	rcCmd := exec.Command("/proc/self/exe", rcArgs...)
	// 设置标准输出输入
	rcCmd.Stderr = os.Stderr
	rcCmd.Stdout = os.Stdout
	rcCmd.Stdin = os.Stdin

	// rootcontainer有自己的uts namespace
	rcCmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS,
	}

	var err error
	if err = rcCmd.Start(); err != nil {
		logger.Fatal("startRootContainerAction Start failed! reason:%s", err.Error())
	}

	logger.Printf("root container pid:%d\n", rcCmd.Process.Pid)

	if err = rcCmd.Wait(); err != nil {
		logger.Fatal("startRootContainerAction Wait failed! reason:%s", err.Error())
	}

	return nil
}

func rootContainerAction(c *cli.Context) error {
	count := c.Int("count")
	logger.Printf("root Container count[%d]\n", count)

	// 设置hostname
	if err := syscall.Sethostname([]byte("InheritedUTS")); err != nil {
		logger.Fatal("Sethostname failed! reason: ", err.Error())
	}

	utsNamespace, _ := os.Readlink("/proc/self/ns/uts")
	hostName, _ := os.Hostname()
	logger.Printf("root Container pid:%d uts namespace:%s, hostName:%s\n", os.Getpid(), utsNamespace, hostName)

	var wg sync.WaitGroup
	// 启动子容器
	for i := 0; i < count; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			ccArgs := []string{"child"}
			ccCmd := exec.Command("/proc/self/exe", ccArgs...)
			// 设置标准输出输入
			ccCmd.Stderr = os.Stderr
			ccCmd.Stdout = os.Stdout
			ccCmd.Stdin = os.Stdin

			var err error
			if err = ccCmd.Start(); err != nil {
				logger.Fatal("startChildContainerAction Start failed! reason:%s", err.Error())
			}

			childContainerPid := ccCmd.Process.Pid
			logger.Printf("start child container pid:%d\n", childContainerPid)

			// 不wait也可以，反正init也可以领取
			// if err = ccCmd.Wait(); err != nil {
			// 	logger.Fatal("startChildContainerAction Wait failed! reason:%s", err.Error())
			// }
			// logger.Printf("child container pid:%d exit!\n", childContainerPid)
		}()
	}

	wg.Wait()
	logger.Printf("root container pid:%d exit!\n", os.Getpid())
	return nil
}

func childContainerAction(c *cli.Context) error {
	time.Sleep(time.Second * 3)
	utsNamespace, _ := os.Readlink("/proc/self/ns/uts")
	hostName, _ := os.Hostname()
	logger.Printf("child Container pid:%d, ppid:%d uts namespace:%s, hostName:%s\n",
		os.Getpid(), os.Getppid(), utsNamespace, hostName)
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "rootcontainer"
	app.Usage = "root container uts namespace test"
	app.Commands = appCmds

	app.Run(os.Args)
}
