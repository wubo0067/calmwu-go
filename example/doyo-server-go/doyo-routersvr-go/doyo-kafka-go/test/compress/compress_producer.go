/*
 * @Author: calmwu
 * @Date: 2018-09-18 10:14:18
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-09-18 10:41:28
 */

package main

import (
	base "doyo-server-go/doyo-base-go"
	"doyo-server-go/doyo-routersvr-go/doyo-kafka-go/doyokafka"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap/zapcore"
)

var (
	ParamMsgCount = flag.Int("count", 100, "")
	ParamMsgSize  = flag.Int("size", 1024, "")
)

func main() {
	flag.Parse()

	base.InitDefaultZapLog("./compress.log", zapcore.DebugLevel)
	km, err := doyokafka.InitModule("10.1.41.149:9094,10.1.41.149:9092,10.1.41.149:9093", []string{}, "1", base.ZLog)
	if err != nil {
		base.ZLog.Errorf("doyokafka InitModule failed: %s", err.Error())
		return
	}

	base.SeedMathRand()
	msg, _ := base.RandomBytes(*ParamMsgSize)
	fmt.Printf("msg size:%d count:%d\n", len(msg), *ParamMsgCount)

	start := time.Now()

	for i := 0; i < *ParamMsgCount; i++ {
		km.PushKfkData("test5", msg)
	}

	fmt.Println(time.Since(start))

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	<-sigchan

	km.ShutDown()
	fmt.Println("compress exit!")
}
