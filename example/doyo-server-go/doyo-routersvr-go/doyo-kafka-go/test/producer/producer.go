/*
 * @Author: calmwu
 * @Date: 2018-09-17 15:47:50
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-09-17 16:19:14
 */

package main

import (
	"bufio"
	base "doyo-server-go/doyo-base-go"
	"doyo-server-go/doyo-routersvr-go/doyo-kafka-go/doyokafka"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap/zapcore"
)

func main() {
	base.InitDefaultZapLog("./producer.log", zapcore.DebugLevel)
	km, err := doyokafka.InitModule("10.1.41.149:9094,10.1.41.149:9092,10.1.41.149:9093", []string{}, "1", base.ZLog)
	if err != nil {
		base.ZLog.Errorf("doyokafka InitModule failed: %s", err.Error())
		return
	}

	consolReader := bufio.NewReader(os.Stdin)
L:
	for {
		fmt.Print("-> ")
		text, _ := consolReader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)

		if strings.Compare("exit", text) == 0 {
			break L
		} else {
			km.PushKfkData("test5", []byte(text))
		}
	}

	fmt.Println("producer exit!")
}
