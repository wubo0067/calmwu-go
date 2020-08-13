/*
 * @Author: calmwu
 * @Date: 2017-12-05 14:38:29
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-12-08 18:59:20
 * @Comment:
 */
package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/wubo0067/calmwu-go/network/transport"
	"github.com/wubo0067/calmwu-go/utils"
)

func main() {
	f, err := os.OpenFile("server.prof", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	err = pprof.StartCPUProfile(f)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer f.Close()
	defer pprof.StopCPUProfile()

	runtime.GOMAXPROCS(runtime.NumCPU())
	utils.InitLog("transport.log")
	defer utils.ZLog.Close()

	config := transport.NewDefaultNetTransportConfig()
	listenIP := "10.10.81.214"
	listenPort := 1003

	tp, err := transport.StartNetTransport(listenIP, listenPort, config)
	if err != nil {
		utils.ZLog.Errorf(err.Error())
		return
	}

	delay := time.After(300 * time.Second)
L:
	for {
		select {
		case netSessionData := <-tp.ReadDataCh():
			utils.ZLog.Debug("netSessionData cmd[%s] sessionid[%d]", netSessionData.Cmd.String(), netSessionData.SessionID)
			if netSessionData.Cmd == transport.E_SESSIONCMD_TRANSFER {
				utils.ZLog.Debug("%s", string(netSessionData.Data))

				data := new(transport.NetSessionData)
				data.Cmd = transport.E_SESSIONCMD_TRANSFER
				data.SessionID = netSessionData.SessionID
				data.Data = []byte("hello client! i am server!")
				tp.WriteData(data)
			}
			transport.PoolPutSessionData(netSessionData)

		case <-delay:
			tp.ShutDown()
			break L
		}
	}
	return
}
