/*
 * @Author: calmwu
 * @Date: 2018-09-15 18:23:21
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-10-29 11:05:13
 */

package main

import (
	base "doyo-server-go/doyo-base-go"
	"doyo-server-go/doyo-routersvr-go/doyo-kafka-go/doyokafka"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap/zapcore"
)

func main() {
	base.InitDefaultZapLog("./consumer.log", zapcore.DebugLevel)
	km, err := doyokafka.InitModule("10.1.41.149:9094,10.1.41.149:9092,10.1.41.149:9093", []string{"test5"}, "1", base.ZLog)
	if err != nil {
		base.ZLog.Errorf("doyokafka InitModule failed: %s", err.Error())
		return
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

L:
	for {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			// 停止从kafka拉取数据
			km.StopPull()
		case data := <-km.PullChan():
			switch d := data.(type) {
			case *doyokafka.DoyoKafkaEofData:
				base.ZLog.Info("receive end notify, The data has been read all!")
				break L
			case *doyokafka.DoyoKafkaReadData:
				base.ZLog.Infof("receive topicInfo:%s, dataSize: %d", d.FromInfo().String(), len(d.Data()))
			}

		}
	}

	km.ShutDown()

	fmt.Println("consumer exit!")
}
