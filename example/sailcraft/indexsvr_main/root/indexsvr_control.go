/*
 * @Author: calmwu
 * @Date: 2017-09-30 10:48:56
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-09-30 10:59:39
 */

package root

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sailcraft/base"
	"sailcraft/indexsvr_main/common"
	"sailcraft/indexsvr_main/data"
	"sailcraft/indexsvr_main/proto"
	"strconv"
	"strings"
	"time"
)

func InitCtrl() error {
	ctrlListener, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP(common.GServListenIP),
		Port: common.GServListenCtrlPort})

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return err
	}

	base.GLog.Debug("Control Interface watch[%s:%d]", common.GServListenIP, common.GServListenCtrlPort)
	go processControlRoutine(ctrlListener)
	return nil
}

func processControlRoutine(listener *net.UDPConn) {
	defer listener.Close()

	base.GLog.Debug("Control routine running")
	cmdBuf := make([]byte, 2048)

	for {
		n, remoteAddr, err := listener.ReadFromUDP(cmdBuf)

		if err != nil {
			base.GLog.Error("Control recv error[%s]", err.Error())
		} else {
			// 处理命令
			realBuf := cmdBuf[:n]
			//base.GLog.Debug("Control recv cmd[%s]", realBuf)
			ctrlCmd := new(proto.ProtoControlCmdS)
			err := json.Unmarshal(realBuf, ctrlCmd)
			if err != nil {
				base.GLog.Error("Control Unmarshal failed! reason[%s]", err.Error())
			} else {
				base.GLog.Debug("%+v", ctrlCmd)
				switch ctrlCmd.CmdName {
				case proto.CTRLCMD_REOLADCONF_REQ:
					base.GLog.Warn("Control recv CTRLCMD_REOLADCONF_REQ")

				case proto.CTRLCMD_RELOADDATA_REQ:
					base.GLog.Info("Control recv CTRLCMD_RELOADDATA_REQ")
					err = data.GDataMgr.Reload()
					result := "OK"
					if err != nil {
						result = err.Error()
					}
					var ctrlCmd proto.ProtoControlCmdS
					ctrlCmd.CmdName = proto.CTRLCMD_RELOADDATA_RES
					ctrlCmd.CmdData = fmt.Sprintf("MonkeyKing[%s:%d] result[%s]",
						common.GServListenIP, common.GServListenCtrlPort, result)
					resBuf, _ := json.Marshal(ctrlCmd)
					listener.WriteToUDP(resBuf, remoteAddr)
				default:
					base.GLog.Error("Control Cmd[%s] is not support!", ctrlCmd.CmdName)
				}
			}
		}
	}
}

func sendControlCmd(addr string, port int, realCmd []byte) (res string) {
	ip := net.ParseIP(addr)

	localAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	remoteAddr := &net.UDPAddr{IP: ip, Port: port}

	conn, err := net.DialUDP("udp", localAddr, remoteAddr)
	if err != nil {
		res = err.Error()
		return
	} else {
		defer conn.Close()
		_, err := conn.Write(realCmd)
		if err != nil {
			res = err.Error()
		} else {
			//fmt.Printf("n[%d] cmdBuf[%s]\n", n, realCmd)
			cmdBuf := make([]byte, 2048)
			// 设置超时时间
			conn.SetReadDeadline(time.Now().Add(time.Second * 5))
			n, _, err := conn.ReadFromUDP(cmdBuf)
			if err != nil {
				res = err.Error()
			} else {
				res = string(cmdBuf[:n])
			}
		}
	}
	return
}

func ReloadCmd(ipNames string, ports string, cmd string) {
	var reloadReqCmd proto.ProtoControlCmdS
	reloadReqCmd.CmdName = cmd
	reloadReqCmd.CmdData = "null"
	//fmt.Printf("%+v\n", reloadReqCmd)
	realCmd, err := json.Marshal(reloadReqCmd)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		//fmt.Println(realCmd)
		ipLst := strings.Split(ipNames, ",")
		portLst := strings.Split(ports, ",")
		for _, ip := range ipLst {
			for _, portName := range portLst {
				port, err := strconv.Atoi(portName)
				if err == nil {
					fmt.Printf("%s[%s:%d] ReloadReq\n", common.GServName, ip, port)
					res := sendControlCmd(ip, port, realCmd)
					fmt.Printf("%s[%s:%d] ReloadRes[%s]\n", common.GServName, ip, port, res)
				} else {
					base.GLog.Error("portName[%s] convert to int failed! reason[%s]", portName, err.Error())
				}
			}
		}
	}
}
