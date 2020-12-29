package main

import (
	"bytes"
	"encoding/binary"
	"kp_proto"
	"net"
	"runtime"
	"strings"
	"sync"
	"time"
	//"fmt"
	//"strconv"
	//"syscall"
)

// 收发应该放到两个独立的goroutine中

type ConnState int

const (
	E_STATE_CONNECTING ConnState = iota
	E_STATE_REGISTERING
	E_STATE_CONNECTED
	E_STATE_DISCONNECTED
	E_STATE_REGISTER_FAILED
	E_STATE_INIT
)

const (
	E_MAX_NETCOMMUNICATION_IDLE_SECS = 30
	E_HEARTBEAT_INTERVAL_SECS        = 10
)

func (s ConnState) String() string {
	switch s {
	case E_STATE_CONNECTING:
		return "E_STATE_CONNECTING"
	case E_STATE_REGISTERING:
		return "E_STATE_REGISTERING"
	case E_STATE_REGISTER_FAILED:
		return "E_STATE_REGISTER_FAILED"
	case E_STATE_CONNECTED:
		return "E_STATE_CONNECTED"
	case E_STATE_DISCONNECTED:
		return "E_STATE_DISCONNECTED"
	default:
		return "E_STATE_INIT"
	}
}

type KPServerInfo struct {
	m_server_ip   string
	m_server_port string
	m_connstate   ConnState
	m_recv_time   int64 // 接受数据的时间，秒
}

func (kpserver *KPServerInfo) final_state() {
	kpserver.m_connstate = E_STATE_DISCONNECTED
}

type KPServerSessions struct {
	m_monitor *sync.Mutex // 连接计数保护
	m_count   int         // 连接计数
}

func (r *KPServerSessions) add_session() int {
	r.m_monitor.Lock()
	count := r.m_count
	r.m_count++
	g_log.Debug("Current count of session is: %d", r.m_count)
	r.m_monitor.Unlock()
	return count
}

func (r *KPServerSessions) dec_session() {
	r.m_monitor.Lock()
	r.m_count--
	g_log.Debug("Current count of session is: %d", r.m_count)
	r.m_monitor.Unlock()
}

var (
	kpserver_map map[string]*KPServerInfo
	g_kpsessions *KPServerSessions = nil
)

func init() {
	if g_kpsessions == nil {
		g_kpsessions = new(KPServerSessions)
		g_kpsessions.m_monitor = new(sync.Mutex)
		g_kpsessions.m_count = 0
	}
}

func init_agent_net(server_infos string) int {
	kpserver_map = make(map[string]*KPServerInfo)

	kpserver_hosts := strings.Split(server_infos, ",")
	g_log.Debug("kpserver_hosts:%v", kpserver_hosts)

	for _, host := range kpserver_hosts {
		g_log.Debug("host:[%s]", host)
		if len(host) == 0 {
			g_log.Critical("host[%s] is invalid!", host)
			return -1
		}

		host_info := strings.Split(host, ":")
		if len(host_info) != 2 {
			g_log.Critical("host[%s] is invalid!", host)
			return -1
		}

		host_ip := host_info[0]
		host_port := host_info[1]

		kpserver_map[host_ip] = &KPServerInfo{
			m_server_ip:   host_ip,
			m_server_port: host_port,
			m_connstate:   E_STATE_INIT,
		}
	}
	return 0
}

func start_agent(kpserver_info *KPServerInfo) {
	go net_process(kpserver_info)
}

func net_process(kpserver_info *KPServerInfo) {
	// 网络处理
	wait_group.Add(1)

	defer wait_group.Done()
	defer kpserver_info.final_state()

	remote_addr := kpserver_info.m_server_ip + ":" + kpserver_info.m_server_port
	g_log.Debug("goroutine net_process[%s] running!", remote_addr)

	// 地址解析
	r_addr, err := net.ResolveTCPAddr("tcp", remote_addr)
	if err != nil {
		g_log.Error("net_process[%s] resolve tcp addr failed! reason[%s]", remote_addr, err.Error())
	}

	kpserver_info.m_connstate = E_STATE_CONNECTING
	session, err := net.DialTCP("tcp", nil, r_addr)
	if err != nil {
		g_log.Error("net_process[%s] connect to kpserver failed! reason[%s]", remote_addr, err.Error())
	} else {
		g_log.Debug("net_process[%s] connect kpserver successed!", remote_addr)
		current_session_count := g_kpsessions.add_session()
		defer session.Close()
		defer g_kpsessions.dec_session()

		// 发送注册信息
		send_ret := send_register_to_kpserver(session, current_session_count == 0)
		if send_ret < 0 {
			return
		}
		kpserver_info.m_connstate = E_STATE_REGISTERING
		kpserver_info.m_recv_time = time.Now().Unix()

		// 心跳定时器
		heartbeat_ticker := time.NewTicker(time.Second * E_HEARTBEAT_INTERVAL_SECS)
		defer heartbeat_ticker.Stop()
		// 尼玛，这是设置Nagle's algorithm，不是nonblock，有了goroutine还要个毛线的nonblock
		session.SetNoDelay(true)

		//read_buf := make([]byte, 1024)
		// 输入缓冲区
		var incmd_buffer *bytes.Buffer = new(bytes.Buffer)
		g_log.Debug("incmd_buffer len[%d] cap[%d]", incmd_buffer.Len(), incmd_buffer.Cap())
	L:
		for {
			// 设置读取超时时间
			session.SetReadDeadline(time.Now().Add(time.Millisecond * 10))
			// 读取数据
			//len, err := session.Read(read_buf)
			ret, err := incmd_buffer.ReadFrom(session)
			//g_log.Debug("goroutine net_process[%s] ret[%d]", remote_addr, ret)

			var is_read_timeout bool = false
			if err != nil {
				// 从接口转换为实际类型
				if nerr, ok := err.(net.Error); !ok || !nerr.Timeout() {
					// 记录出超时之外的所有错误
					g_log.Warn("net_process[%s] read failed! reason[%s]", remote_addr, err.Error())
					// if err == io.EOF {
					// 	// 连接关闭
					// 	g_log.Warn("net_process[%s] active close session!", remote_addr)
					// }
					// 退出
					break L
				} else {
					is_read_timeout = true
				}
			}

			if ret > 0 || incmd_buffer.Len() > 0 {
				// 读取数据插入cmd_buffer中
				g_log.Debug("net read [%d] bytes, incmd_buffer{len:[%d] cap:[%d]}", ret, incmd_buffer.Len(), incmd_buffer.Cap())
				ret := process_data(session, incmd_buffer, kpserver_info, remote_addr)
				if ret < 0 {
					// 数据错误，断开连接
					break L
				} else {
					kpserver_info.m_recv_time = time.Now().Unix()
					// 处理剩余数据
					remain_bytes := incmd_buffer.Bytes()
					if len(remain_bytes) > 0 {
						incmd_buffer.Truncate(0)
						incmd_buffer.Write(remain_bytes)
						g_log.Debug("incmd_buffer len[%d] cap[%d]", incmd_buffer.Len(), incmd_buffer.Cap())
					}
				}
			} else if ret == 0 && !is_read_timeout {
				// 不是超时错误，是读到eof，断开连接
				g_log.Warn("net_process[%s] read EOF, passive close session!", remote_addr)
				break L
			}

			select {
			case <-exit_chan:
				g_log.Debug("net_process[%s] receive exit notify!", remote_addr)
				break L
			case <-heartbeat_ticker.C:
				if kpserver_info.m_connstate == E_STATE_CONNECTED {
					send_ret := send_heartbeat_to_kpserver(session)
					if send_ret < 0 {
						g_log.Error("send heartbeat failed!")
						break L
					}
				}

				if time.Now().Unix()-kpserver_info.m_recv_time >= E_MAX_NETCOMMUNICATION_IDLE_SECS {
					g_log.Warn("session idle timeout! current status[%s]", kpserver_info.m_connstate.String())
					break L
				}
			case rsstatus, ok := <-rsstatus_channel:
				if ok && kpserver_info.m_connstate == E_STATE_CONNECTED {
					send_rsstatus_to_kpserver(session, rsstatus)
				}
			case netstatisinfo, ok := <-netstatis_channel:
				if ok && kpserver_info.m_connstate == E_STATE_CONNECTED {
					send_netstatisinfo_to_kpserver(session, netstatisinfo)
				}
			default:
				continue
			}
		}
	}
	runtime.GC()
	g_log.Debug("goroutine net_process[%s] exit!", remote_addr)
}

func process_data(session *net.TCPConn, incmd_buffer *bytes.Buffer, kpserver_info *KPServerInfo, remote_addr string) int {
	// 判断数据长度
	incmd_buffer_size := incmd_buffer.Len()

	if incmd_buffer_size > 8 {
		// 读取头大小，网络字节序
		pkg_size := binary.BigEndian.Uint32(incmd_buffer.Bytes())
		if incmd_buffer_size >= int(pkg_size) {
			// 获取一个完整包
			cmd_buffer := incmd_buffer.Next(int(pkg_size))
			// for index, c := range cmd_buffer {
			// 	g_log.Debug("cmd_buffer index[%d]\t%d", index, c)
			// }

			// 明文的长度
			plaintext_len := binary.BigEndian.Uint32(cmd_buffer[4:])

			g_log.Debug("package size[%d] plaintext_len[%d] incmd_buffer_size[%d]", pkg_size, plaintext_len, incmd_buffer_size)

			real_data := cmd_buffer[8:]
			if kpserver_info.m_connstate == E_STATE_CONNECTED {
				// 解密
				plaintext, ret := decrypt_ciphertext(real_data)
				if ret != 0 {
					// 解密失败
					return -1
				}
				real_data = plaintext[:plaintext_len]
			}
			// 开始解包
			// for index, c := range real_data {
			// 	g_log.Debug("index[%d]\t%d", index, c)
			// }
			kp_msg, ret := unpack_kpmessage(real_data)
			if ret < 0 {
				return -1
			}
			// 处理包
			ret = process_kpmsg(session, kp_msg, kpserver_info)
			if ret < 0 {
				return -1
			}
		}
	} else {
		// 收到部分包
		g_log.Debug("net_process[%s] recevie part package from kpserver", remote_addr)
	}
	return 0
}

func process_kpmsg(session *net.TCPConn, kp_msg *kp_proto.KPMessage, kpserver_info *KPServerInfo) int {

	g_log.Debug("receive kp_msg[%s] from kpserver[%s]", kp_msg.Cmd.String(), session.RemoteAddr().String())

	switch *kp_msg.Cmd {
	case kp_proto.KPProtoCmd_E_S2A_HEARTBEAT_RES:
		var heartbeat_res kp_proto.HeartBeat
		unpack_kpcmd(kp_msg.RealMessageMarshaldata, &heartbeat_res)

	case kp_proto.KPProtoCmd_E_S2A_REGISTER_RES:
		var register_res kp_proto.S2ARegisterRes
		ret := unpack_kpcmd(kp_msg.RealMessageMarshaldata, &register_res)
		if ret == 0 && *register_res.Result == kp_proto.KPErrorCode_E_KPERR_OK && host_localIP == *register_res.KpLocalip {
			kpserver_info.m_connstate = E_STATE_CONNECTED
			// 计算加密key
			ret = generate_key(register_res.DhKeyServer)
			if ret < 0 {
				return -1
			}

			// 加解密测试
			//test_crypto()
			// 设置注册信息
			g_kpconfinfo.set_registerinfo(&register_res)
		} else {
			g_log.Error("kpagent regsiter failed! reason[%s]", register_res.Result.String())
			return -1
		}

	case kp_proto.KPProtoCmd_E_S2A_ADD_LISTENER_REQ:
		var addlistener_req kp_proto.S2AAddListenerReq
		var addlistener_res kp_proto.A2SAddListenerRes

		ret := unpack_kpcmd(kp_msg.RealMessageMarshaldata, &addlistener_req)
		err_code := kp_proto.KPErrorCode_E_KPERR_OK
		if ret == 0 {
			err_code = g_kpconfinfo.add_virtual_serv(&addlistener_req)
		} else {
			g_log.Error("unpack KPProtoCmd_E_S2A_ADD_LISTENER_REQ msg failed!")
			return -1
		}
		addlistener_res.ProxyId = &g_kpconfinfo.m_proxyid
		addlistener_res.VrrpRouterid = &g_kpconfinfo.VrrpRouterID
		addlistener_res.ListenerId = addlistener_req.ListenerId
		addlistener_res.TaskId = addlistener_req.TaskId
		addlistener_res.NetProtocolName = addlistener_req.NetProtocolName
		addlistener_res.ListenerPort = addlistener_req.ListenerPort
		addlistener_res.Result = &err_code
		marshal_data, ret := pack_kpcmd(&addlistener_res)
		if ret == 0 {
			return send_kpmsg(session, kp_proto.KPProtoCmd_E_A2S_ADD_LISTENER_RES, marshal_data, true)
		}
		return -1

	case kp_proto.KPProtoCmd_E_S2A_DEL_LISTENER_REQ:
		var dellistener_req kp_proto.S2ADelListenerReq
		var dellistener_res kp_proto.A2SDelListenerRes

		err_code := kp_proto.KPErrorCode_E_KPERR_OK
		ret := unpack_kpcmd(kp_msg.RealMessageMarshaldata, &dellistener_req)
		if ret == 0 {
			err_code = g_kpconfinfo.del_virtual_serv(&dellistener_req)
		} else {
			g_log.Error("unpack KPProtoCmd_E_S2A_DEL_LISTENER_REQ msg failed!")
			return -1
		}

		dellistener_res.ProxyId = &g_kpconfinfo.m_proxyid
		dellistener_res.TaskId = dellistener_req.TaskId
		dellistener_res.VrrpRouterid = &g_kpconfinfo.VrrpRouterID
		dellistener_res.ListenerId = dellistener_req.ListenerId
		dellistener_res.NetProtocolName = dellistener_req.NetProtocolName
		dellistener_res.ListenerPort = dellistener_req.ListenerPort
		dellistener_res.Result = &err_code
		marshal_data, ret := pack_kpcmd(&dellistener_res)
		if ret == 0 {
			return send_kpmsg(session, kp_proto.KPProtoCmd_E_A2S_DEL_LISTENER_RES, marshal_data, true)
		}
		return -1

	case kp_proto.KPProtoCmd_E_S2A_BIND_LISTENER_WITH_RS_REQ:
		var bindrs_req kp_proto.S2ABindListenerWithRSReq
		var bindrs_res kp_proto.A2SBindListenerWithRSRes

		err_code := kp_proto.KPErrorCode_E_KPERR_OK
		ret := unpack_kpcmd(kp_msg.RealMessageMarshaldata, &bindrs_req)
		if ret == 0 {
			err_code = g_kpconfinfo.bind_virtualsvr_realsvr(&bindrs_req)
		} else {
			g_log.Error("unpack KPProtoCmd_E_S2A_BIND_LISTENER_WITH_RS_REQ msg failed!")
			return -1
		}
		bindrs_res.ProxyId = &g_kpconfinfo.m_proxyid
		bindrs_res.TaskId = bindrs_req.TaskId
		bindrs_res.VrrpRouterid = &g_kpconfinfo.VrrpRouterID
		bindrs_res.ListenerId = bindrs_req.ListenerId
		bindrs_res.NetProtocolName = bindrs_req.NetProtocolName
		bindrs_res.ListenerPort = bindrs_req.ListenerPort
		bindrs_res.Result = &err_code
		marshal_data, ret := pack_kpcmd(&bindrs_res)
		if ret == 0 {
			return send_kpmsg(session, kp_proto.KPProtoCmd_E_A2S_BIND_LISTENER_WITH_RS_RES, marshal_data, true)
		}
		return -1

	case kp_proto.KPProtoCmd_E_S2A_UNBIND_LISTENER_WITH_RS_REQ:
		var unbindrs_req kp_proto.S2AUnBindListenerWithRsReq
		var unbindrs_res kp_proto.A2SUnBindListenerWithRSRes

		err_code := kp_proto.KPErrorCode_E_KPERR_OK
		ret := unpack_kpcmd(kp_msg.RealMessageMarshaldata, &unbindrs_req)
		if ret == 0 {
			err_code = g_kpconfinfo.unbind_virtualsvr_realsvr(&unbindrs_req)
		} else {
			g_log.Error("unpack KPProtoCmd_E_S2A_UNBIND_LISTENER_WITH_RS_REQ msg failed!")
			return -1
		}

		unbindrs_res.ProxyId = &g_kpconfinfo.m_proxyid
		unbindrs_res.TaskId = unbindrs_req.TaskId
		unbindrs_res.VrrpRouterid = &g_kpconfinfo.VrrpRouterID
		unbindrs_res.ListenerId = unbindrs_req.ListenerId
		unbindrs_res.NetProtocolName = unbindrs_req.NetProtocolName
		unbindrs_res.ListenerPort = unbindrs_req.ListenerPort
		unbindrs_res.Result = &err_code
		marshal_data, ret := pack_kpcmd(&unbindrs_res)

		if ret == 0 {
			return send_kpmsg(session, kp_proto.KPProtoCmd_E_A2S_UNBIND_LISTENER_WITH_RS_RES, marshal_data, true)
		}
		return -1

	case kp_proto.KPProtoCmd_E_S2A_MOD_LISTENER_REQ:
		var modlistener_req kp_proto.S2AModListenerReq
		var modlistener_res kp_proto.A2SModListenerRes

		ret := unpack_kpcmd(kp_msg.RealMessageMarshaldata, &modlistener_req)
		if ret == 0 {
			err_code := kp_proto.KPErrorCode_E_KPERR_OK
			if *modlistener_req.ProxyId == g_kpconfinfo.m_proxyid && *modlistener_req.VrrpRouterid == g_kpconfinfo.VrrpRouterID {
				err_code = g_kpconfinfo.update_virtualsvr(&modlistener_req)
			} else {
				g_log.Error("unpack KPProtoCmd_E_S2A_MOD_LISTENER_REQ msg failed!")
				return -1
			}

			modlistener_res.ProxyId = &g_kpconfinfo.m_proxyid
			modlistener_res.TaskId = modlistener_req.TaskId
			modlistener_res.VrrpRouterid = &g_kpconfinfo.VrrpRouterID
			modlistener_res.ListenerId = modlistener_req.ListenerId
			modlistener_res.NetProtocolName = modlistener_req.NetProtocolName
			modlistener_res.ListenerPort = modlistener_req.ListenerPort
			modlistener_res.Result = &err_code

			marshal_data, ret := pack_kpcmd(&modlistener_res)
			if ret == 0 {
				return send_kpmsg(session, kp_proto.KPProtoCmd_E_A2S_MOD_LISTENER_RES, marshal_data, true)
			}
			return -1
		}
		return -1
	case kp_proto.KPProtoCmd_E_S2A_KPRELOAD_REQ:
		var reload_req kp_proto.S2AReloadReq
		var reload_res kp_proto.A2SReloadRes

		ret := unpack_kpcmd(kp_msg.RealMessageMarshaldata, &reload_req)
		if ret == 0 {
			if *reload_req.ProxyId == g_kpconfinfo.m_proxyid && *reload_req.VrrpRouterid == g_kpconfinfo.VrrpRouterID {
				// 执行keepalived reload操作
				var err_code kp_proto.KPErrorCode = kp_proto.KPErrorCode_E_KPERR_OK
				var err_reason string = ""
				err_code, err_reason = g_kpconfinfo.reload_keepalived()
				reload_res.ProxyId = &g_kpconfinfo.m_proxyid
				reload_res.VrrpRouterid = &g_kpconfinfo.VrrpRouterID
				reload_res.TaskId = reload_req.TaskId
				reload_res.Result = &err_code
				reload_res.FailedReason = &err_reason
				marshal_data, ret := pack_kpcmd(&reload_res)
				if ret == 0 {
					return send_kpmsg(session, kp_proto.KPProtoCmd_E_A2S_KPRELOAD_RES, marshal_data, true)
				}
			} else {
				g_log.Error("ProxyId[%s] VrrpRouterid[%d] do not match local setting")
			}
		}
		return -1

	default:
		g_log.Error("receive cmd is Unknown!", *kp_msg.Cmd)
	}
	return 0
}

func send_register_to_kpserver(session *net.TCPConn, is_master bool) int {
	var register_req kp_proto.A2SRegisterReq
	register_req.AgentLocalip = &host_localIP
	register_req.DhKeyAgent = get_dhka()
	register_req.IsMaster = &is_master
	marshal_data, ret := pack_kpcmd(&register_req)
	if ret == 0 {
		return send_kpmsg(session, kp_proto.KPProtoCmd_E_A2S_REGISTER_REQ, marshal_data, false)
	}
	return -1
}

func send_heartbeat_to_kpserver(session *net.TCPConn) int {
	var heartbeat_req kp_proto.HeartBeat
	time_stamp := time.Now().Unix()
	heartbeat_req.TimeStamp = &time_stamp
	heartbeat_req.FromHost = &host_localIP
	// 这里其实传入的是指针
	marshal_data, ret := pack_kpcmd(&heartbeat_req)
	if ret == 0 {
		return send_kpmsg(session, kp_proto.KPProtoCmd_E_A2S_HEARTBEAT_REQ, marshal_data, true)
	}
	return -1
}

func send_rsstatus_to_kpserver(session *net.TCPConn, rsstatus kp_proto.RSStatus) int {
	var rsstatus_ntf kp_proto.A2SRSStatusNotify
	rsstatus_ntf.ProxyId = &g_kpconfinfo.m_proxyid
	rsstatus_ntf.VrrpRouterid = &g_kpconfinfo.VrrpRouterID
	rsstatus_ntf.StatusInfo = &rsstatus
	ret, listenerid := g_kpconfinfo.find_listenerid(rsstatus.RsPort, rsstatus.NetProtocolName)
	if ret == 0 {
		rsstatus_ntf.StatusInfo.ListenerId = &listenerid
	}
	marshal_data, ret := pack_kpcmd(&rsstatus_ntf)
	if ret == 0 {
		return send_kpmsg(session, kp_proto.KPProtoCmd_E_A2S_RS_STATUS_NOTIFY, marshal_data, true)
	}
	return -1
}

func send_netstatisinfo_to_kpserver(session *net.TCPConn, netstatisinfo kp_proto.NetStatisticsInfo) int {
	var netstatisinfo_ntf kp_proto.A2SNetStatisticsNtf
	netstatisinfo_ntf.ProxyId = &g_kpconfinfo.m_proxyid
	netstatisinfo_ntf.VrrpRouterid = &g_kpconfinfo.VrrpRouterID
	netstatisinfo_ntf.LinkNodeId = &g_kpconfinfo.m_linknodeid
	netstatisinfo_ntf.Netstatis = &netstatisinfo
	time_stamp := time.Now().Unix()
	netstatisinfo_ntf.TimeStamp = &time_stamp
	marshal_data, ret := pack_kpcmd(&netstatisinfo_ntf)
	if ret == 0 {
		return send_kpmsg(session, kp_proto.KPProtoCmd_E_A2S_NETSTATISTICS_NTF, marshal_data, true)
	}
	return -1
}

func send_kpmsg(session *net.TCPConn, cmd kp_proto.KPProtoCmd, cmd_body []byte, is_encrypt bool) int {
	// 打包
	cmd_marshal_data, ret := pack_kpmessage(cmd, cmd_body)
	if ret < 0 {
		return -1
	}

	ciphertext_buff := cmd_marshal_data
	ciphertext_len := len(cmd_marshal_data)
	if is_encrypt {
		// 加密
		ciphertext_buff, ciphertext_len = encrypt_plaintext(cmd_marshal_data)
		if ciphertext_len < 0 {
			return -1
		}
	}

	// 数据总长度 |数据总长度|明文长度|实际数据|
	kp_message_size := ciphertext_len + 8
	// 打包的缓冲区
	kp_message_buff := make([]byte, kp_message_size)

	// 写入数据长度
	//binary.Write(kp_message_buff, binary.BigEndian, &kp_message_size)
	binary.BigEndian.PutUint32(kp_message_buff, uint32(kp_message_size))
	binary.BigEndian.PutUint32(kp_message_buff[4:], uint32(len(cmd_marshal_data)))
	// 写入数据
	copy(kp_message_buff[8:], ciphertext_buff)
	// 发送数据
	ret, err := session.Write(kp_message_buff)
	if err != nil {
		g_log.Error("send kpmessage[%s] to kpserver[%s] failed! reason[%s]", cmd.String(), session.RemoteAddr().String(), err.Error())
		return -1
	}
	g_log.Debug("send kpmessage[%s] to kpserver[%s] bytes[%d] successd", cmd.String(), session.RemoteAddr().String(), ret)

	return 0
}
