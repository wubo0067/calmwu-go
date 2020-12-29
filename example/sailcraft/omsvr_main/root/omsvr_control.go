/*
 * @Author: calmwu
 * @Date: 2018-05-18 11:06:50
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-18 11:07:13
 * @Comment:
 */

package root

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sailcraft/base"
	"sailcraft/omsvr_main/common"
	"strconv"
	"strings"
	"time"
)

const (
	CMD_VERIFY_KEY = "6de7fd14a2fa5fdfb541808745cc4267"
	CMD_STOP_REQ   = "stop_req"
	CMD_STOP_RES   = "stop_res"
	CMD_RELOAD_REQ = "reload_req"
	CMD_RELOAD_RES = "reload_res"
)

type ControlCmdS struct {
	CmdVerifyKey string `json:"CmdVerifyKey"`
	CmdName      string `json:"CmdName"`
	CmdData      string `json:"CmdData"`
}

func InitSvrCtrl(ctrlListenIP string, ctrlListenPort int) error {
	ctrlListener, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP(ctrlListenIP),
		Port: ctrlListenPort})

	if err != nil {
		base.GLog.Error("ListenUDP failed! reason[%s]", err.Error())
		return err
	}

	base.GLog.Debug("Control Interface watch[%s:%d]", ctrlListenIP, ctrlListenPort)
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
			//common.GLog.Debug("Control recv cmd[%s]", realBuf)
			ctrlCmd := new(ControlCmdS)
			err := json.Unmarshal(realBuf, ctrlCmd)
			if err != nil {
				base.GLog.Error("Control Unmarshal failed! reason[%s]", err.Error())
			} else {
				base.GLog.Debug("%+v", ctrlCmd)
				switch ctrlCmd.CmdName {
				case CMD_STOP_REQ:
					base.GLog.Info("OMSvr receive CMD_STOP_REQ")
					os.Exit(1)
				case CMD_RELOAD_REQ:
					base.GLog.Info("OMSvr receive CMD_RELOAD_REQ")
					var reloadResCmd ControlCmdS
					reloadResCmd.CmdName = CMD_RELOAD_RES
					reloadResCmd.CmdData = fmt.Sprintf("OMSvr[%s] result[%s]",
						listener.LocalAddr().String(), common.GConfig.ReloadConfig())
					reloadResCmd.CmdVerifyKey = CMD_VERIFY_KEY
					resBuf, _ := json.Marshal(reloadResCmd)
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

func doReloadCmd(reloadIps string, ports string) {
	var reloadReqCmd ControlCmdS
	reloadReqCmd.CmdVerifyKey = CMD_VERIFY_KEY
	reloadReqCmd.CmdName = CMD_RELOAD_REQ
	reloadReqCmd.CmdData = "null"
	//fmt.Printf("%+v\n", reloadReqCmd)
	realCmd, err := json.Marshal(reloadReqCmd)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		//fmt.Println(realCmd)
		ipLst := strings.Split(reloadIps, ",")
		portLst := strings.Split(ports, ",")
		for _, ip := range ipLst {
			for _, portName := range portLst {
				port, err := strconv.Atoi(portName)
				if err == nil {
					fmt.Printf("OMSvr[%s:%d] ReloadReq\n", ip, port)
					res := sendControlCmd(ip, port, realCmd)
					fmt.Printf("OMSvr[%s:%d] ReloadRes[%s]\n", ip, port, res)
				} else {
					base.GLog.Error("portName[%s] convert to int failed! reason[%s]", portName, err.Error())
				}
			}
		}
	}
}
