package main

import (
	"kp_proto"
	"net"
	"strings"
	"time"
	"unsafe"
)

var (
	// rs状态通知的channel
	rsstatus_channel = make(chan kp_proto.RSStatus, 1024)
)

func notify_rs_status() {
	g_log.Debug("rs_ip[%s] rs_port[%d] proto_name[%s] status[%s]",
		*cmd_params_rs_ip, *cmd_params_rs_port, *cmd_params_proto_name, *cmd_params_notify_status)

	ip := net.ParseIP("127.0.0.1")
	var ListenerID string = ""

	local_addr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	remote_addr := &net.UDPAddr{IP: ip, Port: *cmd_params_notify_port}

	conn, err := net.DialUDP("udp", local_addr, remote_addr)
	if err != nil {
		g_log.Error(err.Error())
	} else {
		defer conn.Close()

		var rs_status kp_proto.RSStatus
		rs_status.RsPort = (*int32)(unsafe.Pointer(cmd_params_rs_port))
		rs_status.RsIp = cmd_params_rs_ip
		rs_status.NetProtocolName = cmd_params_proto_name
		rs_status.ListenerId = &ListenerID

		var status int32 = 0
		if strings.Compare(*cmd_params_notify_status, "up") == 0 {
			status = 1
		}
		rs_status.Status = &status

		var now int64 = time.Now().Unix()
		rs_status.TimeStamp = &now

		g_log.Debug("notify rsstatus [%s]", rs_status.String())
		// 打包
		data, ret := pack_kpcmd(&rs_status)
		if ret == 0 {
			// 发送数据
			size, err := conn.Write(data)
			if err != nil {
				g_log.Error(err.Error())
			} else {
				g_log.Debug("send %d bytes to %s successed!", size, remote_addr.String())
			}
		}
	}
}

func process_rsstatus_notify(listener *net.UDPConn) {
	g_log.Debug("goroutine process_rsstatus_notify running")
	var rs_status kp_proto.RSStatus
	notify_data := make([]byte, 1024)
L:
	for {
		listener.SetReadDeadline(time.Now().Add(time.Millisecond * 10))

		n, _, err := listener.ReadFromUDP(notify_data)

		if err != nil {
			if nerr, ok := err.(net.Error); !ok || !nerr.Timeout() {
				// 记录出超时之外的所有错误
				g_log.Warn("rs status read failed! reason[%s]", err.Error())
			}
		} else {
			// 将网络消息发送到tcp通道中
			ret := unpack_kpcmd(notify_data[:n], &rs_status)
			if ret == 0 {
				rsstatus_channel <- rs_status
			}
			g_log.Debug("receive rsstatus [%s]", rs_status.String())
		}

		select {
		case <-exit_chan:
			g_log.Debug("process_rsstatus_notify receive exit notify!")
			break L
		default:
			continue
		}
	}
}
