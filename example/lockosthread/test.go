/*
 * @Author: calm.wu
 * @Date: 2019-11-07 14:13:12
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-11-07 15:29:12
 */

package main

import (
	"fmt"
	"net/http"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/sys/unix"
)

func main() {
	// 主协程绑定一个独立的线程M，其实这就是个线程了，也就是所谓的内核线程
	runtime.LockOSThread()

	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		// 子协程不会继承主协程的LockOSThread
		fmt.Printf("worker tid:%d tid:%d", syscall.Gettid(), unix.Gettid())
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	s := &http.Server{
		Addr:           ":18282",
		Handler:        router,
		ReadTimeout:    600 * time.Second,
		WriteTimeout:   600 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Printf("listen on :18282, tid:%d", syscall.Gettid())
	s.ListenAndServe()
}
