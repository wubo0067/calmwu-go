// +build linux

/*
 * @Author: calmwu
 * @Date: 2017-12-04 17:06:02
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-12-05 15:07:52
 * @Comment:
 */

package transport

import (
	"testing"
	"time"

	"github.com/wubo0067/calmwu-go/utils"
)

func initLog() {

}

func TestStartTransport(t *testing.T) {
	utils.InitLog("transport.log")
	defer utils.ZLog.Close()

	config := NewDefaultNetTransportConfig()
	listenIP := "0.0.0.0"
	listenPort := 1003

	transport, err := StartNetTransport(listenIP, listenPort, config)
	if err != nil {
		t.Error(err.Error())
		return
	}

	time.Sleep(100 * time.Second)

	transport.ShutDown()
}
