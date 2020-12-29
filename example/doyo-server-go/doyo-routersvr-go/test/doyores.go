/*
 * @Author: calmwu
 * @Date: 2018-10-03 11:30:38
 * @Last Modified by:   calmwu
 * @Last Modified time: 2018-10-03 11:30:38
 */

package main

import (
	base "doyo-server-go/doyo-base-go"
	"doyo-server-go/doyo-routersvr-go/doyo-kafka-go/doyokafka"
	"doyo-server-go/doyo-routersvr-go/doyo-routerstub"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pquerna/ffjson/ffjson"

	"go.uber.org/zap/zapcore"
)

var (
	cmdParamSvrName    = flag.String("svrname", "DoyoResApp", "")
	cmdParamID         = flag.Int("id", 1, "")
	cmdParamKfkBrokers = flag.String("brokers", "", "")
	cmdParamHostIP     = flag.String("ip", "127.0.0.1", "")

	svrInstanceTopic string
)

func OnReceive(doyoRsm *routerstub.RouterStubModule, msgID string, fromTopic string, data []byte) {
	base.ZLog.Debugf("msgID[%s] fromTopic[%s] data[%s]", msgID, fromTopic, string(data))

	var appMsg AppMsg
	err := ffjson.Unmarshal(data, &appMsg)
	if err != nil {
		base.ZLog.Errorf("ffjson Unmarshal msgID[%s] fromTopic[%s] failed! reason:%s", msgID, fromTopic, err.Error())
	} else {
		switch appMsg.Cmd {
		case CMD_REQ_REVERSESTRING:
			var reverseStrReq ReverseStringReq
			var reverseStrRes ReverseStringRes

			ffjson.Unmarshal(appMsg.Body, &reverseStrReq)
			reverseStrRes.ReverseString = base.ReverseString(reverseStrReq.NormalString)

			appMsg.Body, _ = ffjson.Marshal(reverseStrRes)
			appMsg.Cmd = CMD_RES_REVERSESTRING
			jsonMsg, _ := ffjson.Marshal(appMsg)
			// 返回给调用方
			routerstub.Reply(doyoRsm, fromTopic, msgID, svrInstanceTopic, jsonMsg)
		case CMD_NTF_HELLO:
			var helloNtf HelloNtf
			ffjson.Unmarshal(appMsg.Body, &helloNtf)
			base.ZLog.Infof("Receive msgID[%s] fromTopic[%s] cmd[%d] helloInfo[%s]", msgID, fromTopic, appMsg.Cmd, helloNtf.HelloInfo)
		}
	}
}

func main() {
	flag.Parse()

	svrInstanceTopic = fmt.Sprintf("%s-%d", *cmdParamSvrName, *cmdParamID)

	base.InitDefaultZapLog(fmt.Sprintf("%s.log", svrInstanceTopic), zapcore.DebugLevel)

	doyoKfk, err := doyokafka.InitModule(*cmdParamKfkBrokers, []string{svrInstanceTopic}, svrInstanceTopic, base.ZLog)
	if err != nil {
		base.ZLog.Errorf("doyokafka InitModule failed: %s", err.Error())
		return
	}

	doyoRsm, err := routerstub.NewRouterStubModule(*cmdParamSvrName, svrInstanceTopic, doyoKfk, OnReceive,
		*cmdParamHostIP, (4500 + *cmdParamID))
	if err != nil {
		base.ZLog.Errorf("NewRouterStubModule failed, reason:%s", err.Error())
		return
	}

	// 信号处理
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

L:
	for {
		select {
		case sig := <-sigchan:
			switch sig {
			case syscall.SIGINT:
				fallthrough
			case syscall.SIGTERM:
				base.ZLog.Warnf("Receive Stop Notify sig: %s", sig.String())
				// 停止从kafka拉取数据
				doyoKfk.StopPull()
			case syscall.SIGUSR1:
				base.ZLog.Warnf("Receive Reload Notify")
			}
		case data := <-doyoKfk.PullChan():
			switch d := data.(type) {
			// 从kafka拉取数据结束
			case *doyokafka.DoyoKafkaEofData:
				base.ZLog.Info("receive end notify, The data has been read all!")
				break L
			case *doyokafka.DoyoKafkaReadData:
				doyoRsm.ReceiveDoyoKfkData(d)
			}
		}
	}

	doyoRsm.Stop()
	doyoKfk.ShutDown()
}
