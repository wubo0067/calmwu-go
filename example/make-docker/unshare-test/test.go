/*
 * @Author: calm.wu
 * @Date: 2019-11-08 16:50:03
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-11-10 16:39:19
 */

package main

import (
	"os"
	"runtime"

	"golang.org/x/sys/unix"

	"github.com/vishvananda/netns"
	calm_utils "github.com/wubo0067/calmwu-go/utils"
)

func main() {
	logger := calm_utils.NewSimpleLog(nil)

	logger.Printf("/proc/self/ns/net pid:%d\n", os.Getpid())

	netNsID, err := os.Readlink("/proc/self/ns/net")
	if err != nil {
		logger.Fatalf("Readlink failed. err:%s\n", err.Error())
	}

	logger.Printf("pid:%d /proc/self/ns/net:%s\n", os.Getpid(), netNsID)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	originNs, err := netns.Get()
	if err != nil {
		logger.Fatalf("netns.Get failed. err:%s\n", err.Error())
	}
	defer originNs.Close()

	logger.Printf("originNs:%v\n", originNs)

	err = unix.Unshare(unix.CLONE_NEWNET)
	if err != nil {
		logger.Fatalf("Unshare CLONE_NEWNET failed. err:%s\n", err.Error())
	}

	netNsID, err = os.Readlink("/proc/self/ns/net")
	if err != nil {
		logger.Fatalf("Readlink failed. err:%s\n", err.Error())
	}
	logger.Printf("pid:%d /proc/self/ns/net:%s\n", os.Getpid(), netNsID)
}
