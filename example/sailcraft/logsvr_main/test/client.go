/*
 * @Author: calmwu
 * @Date: 2017-09-02 22:05:47
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-09-02 22:16:15
 */

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"logsvr_main/proto"
	"net"
	"os"
)

var (
	cmdParamsIP    = flag.String("ip", "0.0.0.0", "logsvr server listen ip")
	cmdParamsPort  = flag.Int("port", 5005, "logsvr server listen port")
	cmdParamsTimes = flag.Int("times", 10, "send times")
)

func main() {
	flag.Parse()

	var logInfo proto.ProtoLogInfoS
	logInfo.HostIP = "2.2.3.4"
	logInfo.ServerID = "Skenet-2"
	logInfo.LogLevel = 2
	logInfo.FileName = "client.go"
	logInfo.LineNo = 10
	logInfo.LogContent = "hello! I'm client!"

	marshalData, err := json.Marshal(logInfo)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	local_addr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	remote_addr := &net.UDPAddr{IP: net.ParseIP(*cmdParamsIP), Port: *cmdParamsPort}
	conn, err := net.DialUDP("udp", local_addr, remote_addr)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	} else {
		defer conn.Close()

		for i := 0; i < *cmdParamsTimes; i++ {
			size, err := conn.Write(marshalData)
			if err != nil {
				fmt.Println(err.Error)
				os.Exit(-1)
			} else {
				fmt.Printf("Send to %s [%d] bytes\n", remote_addr.String(), size)
			}
		}
	}
}
