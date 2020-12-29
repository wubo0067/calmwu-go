package main

import (
	//"fmt"
	"kp_proto"
	"strings"
	"sync"
)

type KPRealServer struct {
	RsIP   string
	RsPort int32
	Weight int32 // 只有在策略是wrr的时候权重才有效
}

type KPVritualServer struct {
	NetProtoName   string // tcp udp
	Port           int32  // 监听器的端口
	Scheduler      string // 调度策略
	DelayLoop      int32
	ListenerID     string
	RealSvrInfoLst []*KPRealServer // rs ip列表
}

type KPConfInfo struct {
	m_proxyid         string             //  产品的proxyid
	VrrpRouterID      int32              //
	VrrpInterface     string             //
	LocalIP           string             //
	VirtualIP         string             // 这台代理主机的接入ip
	VirtualSvrInfoLst []*KPVritualServer // VirtualServ列表

	m_linknodeid string
	LinkNodeType kp_proto.LinkNodeType

	m_monitor *sync.Mutex // 数据保护mutex
}

var (
	g_kpconfinfo *KPConfInfo = nil
)

func init() {
	// 初始化
	if g_kpconfinfo == nil {
		g_kpconfinfo = new(KPConfInfo)
		g_kpconfinfo.m_monitor = new(sync.Mutex)
	}
}

func (kpci *KPConfInfo) print_kpci() {
	if kpci != nil {
		g_log.Debug("proxyid[%s] vrrp_routerid[%d] vrrp_interface[%s] KPlocalIP[%s] KPVirtualIP[%s] VirtualServCount[%d] linknodeid[%s] linknodetype[%s]",
			kpci.m_proxyid,
			kpci.VrrpRouterID,
			kpci.VrrpInterface,
			kpci.LocalIP,
			kpci.VirtualIP,
			len(kpci.VirtualSvrInfoLst),
			kpci.m_linknodeid,
			kpci.LinkNodeType.String())

		for index, _ := range kpci.VirtualSvrInfoLst {
			g_log.Debug("%d: ListenerID[%s] protocol[%s] port[%d] rs_lst_count[%d]",
				index,
				kpci.VirtualSvrInfoLst[index].ListenerID,
				kpci.VirtualSvrInfoLst[index].NetProtoName,
				kpci.VirtualSvrInfoLst[index].Port,
				len(kpci.VirtualSvrInfoLst[index].RealSvrInfoLst))

			for rs_index, _ := range kpci.VirtualSvrInfoLst[index].RealSvrInfoLst {
				g_log.Debug("rs:%v", kpci.VirtualSvrInfoLst[index].RealSvrInfoLst[rs_index])
			}
		}
	}
}

func (kpci *KPConfInfo) set_registerinfo(register_info *kp_proto.S2ARegisterRes) {
	if kpci != nil {
		g_log.Debug(register_info.String())
		kpci.m_monitor.Lock()
		defer kpci.m_monitor.Unlock()

		kpci.m_proxyid = *register_info.ProxyId
		kpci.VrrpRouterID = *register_info.VrrpRouterid
		kpci.LocalIP = *register_info.KpLocalip
		kpci.VirtualIP = *register_info.KpVirtualIp
		kpci.VrrpInterface = *register_info.VrrpItfname
		kpci.m_linknodeid = *register_info.LinkNodeId
		kpci.LinkNodeType = *register_info.LinkNodeType
		if len(kpci.VirtualSvrInfoLst) > 0 {
			kpci.VirtualSvrInfoLst = kpci.VirtualSvrInfoLst[0:0]
		}

		if kpci.LinkNodeType != kp_proto.LinkNodeType_E_LINKNODE_FRONTEND {
			exec_close_fullnat_toa_entry()
		}
		kpci.print_kpci()

		g_log.Debug("set register successed!")
	}
}

// 添加监听器
func (kpci *KPConfInfo) add_virtual_serv(virtualserv_addreq *kp_proto.S2AAddListenerReq) kp_proto.KPErrorCode {
	if kpci == nil {
		return kp_proto.KPErrorCode_E_KPERR_FAILED_INSIDERROR
	}
	g_log.Debug(virtualserv_addreq.String())

	kpci.m_monitor.Lock()
	defer kpci.m_monitor.Unlock()

	// 校验proxyid和vrrpid是否相同
	ret := kpci.check_proxyid_and_routerid(*virtualserv_addreq.ProxyId, *virtualserv_addreq.VrrpRouterid)
	if ret != kp_proto.KPErrorCode_E_KPERR_OK {
		return ret
	}

	// 判断virtualsvr是否已经存在
	virtual_serv, _ := kpci.find_virtual_serv(virtualserv_addreq.NetProtocolName, *virtualserv_addreq.ListenerPort)
	if virtual_serv != nil {
		g_log.Error("VirtualServ net_proto_name[%s] listen_port[%d] already exist! update it", *virtualserv_addreq.NetProtocolName,
			*virtualserv_addreq.ListenerPort)
		virtual_serv.NetProtoName = *virtualserv_addreq.NetProtocolName
		virtual_serv.Port = *virtualserv_addreq.ListenerPort
		virtual_serv.Scheduler = *virtualserv_addreq.Scheduler
		virtual_serv.DelayLoop = *virtualserv_addreq.DelayLoop
		virtual_serv.ListenerID = *virtualserv_addreq.ListenerId
		// 清空rs列表
		virtual_serv.RealSvrInfoLst = virtual_serv.RealSvrInfoLst[0:0]
	} else {
		virtual_serv = new(KPVritualServer)
		virtual_serv.NetProtoName = *virtualserv_addreq.NetProtocolName
		virtual_serv.Port = *virtualserv_addreq.ListenerPort
		virtual_serv.Scheduler = *virtualserv_addreq.Scheduler
		virtual_serv.DelayLoop = *virtualserv_addreq.DelayLoop
		virtual_serv.ListenerID = *virtualserv_addreq.ListenerId
		// 添加
		kpci.VirtualSvrInfoLst = append(kpci.VirtualSvrInfoLst, virtual_serv)
	}
	// 如果是中间proxy节点，这里需要有后端的realsvr
	for index, _ := range virtualserv_addreq.ProxyRealsvrLst {
		real_serv := new(KPRealServer)
		real_serv.RsIP = *virtualserv_addreq.ProxyRealsvrLst[index].Ip
		real_serv.Weight = *virtualserv_addreq.ProxyRealsvrLst[index].Weight
		real_serv.RsPort = *virtualserv_addreq.ListenerPort
		virtual_serv.RealSvrInfoLst = append(virtual_serv.RealSvrInfoLst, real_serv)
	}

	kpci.print_kpci()
	return kp_proto.KPErrorCode_E_KPERR_OK
}

// 删除监听器
func (kpci *KPConfInfo) del_virtual_serv(virtualserv_delreq *kp_proto.S2ADelListenerReq) kp_proto.KPErrorCode {
	if kpci == nil {
		return kp_proto.KPErrorCode_E_KPERR_FAILED_INSIDERROR
	}

	kpci.m_monitor.Lock()
	defer kpci.m_monitor.Unlock()

	// 校验proxyid和vrrpid是否相同
	ret := kpci.check_proxyid_and_routerid(*virtualserv_delreq.ProxyId, *virtualserv_delreq.VrrpRouterid)
	if ret != kp_proto.KPErrorCode_E_KPERR_OK {
		return ret
	}

	// 判断virtualsvr是否已经存在
	virtual_serv, pos := kpci.find_virtual_serv(virtualserv_delreq.NetProtocolName, *virtualserv_delreq.ListenerPort)
	if virtual_serv == nil {
		g_log.Error("VirtualServ net_proto_name[%s] listen_port[%d] is not exist!", *virtualserv_delreq.NetProtocolName,
			*virtualserv_delreq.ListenerPort)
		return kp_proto.KPErrorCode_E_KPERR_FAILED_DELVIRTUALSVR_NOTEXISTS
	}

	// 从slice中删除
	kpci.VirtualSvrInfoLst = append(kpci.VirtualSvrInfoLst[:pos], kpci.VirtualSvrInfoLst[pos+1:]...)

	kpci.print_kpci()
	return kp_proto.KPErrorCode_E_KPERR_OK
}

// 监听器绑定rs
func (kpci *KPConfInfo) bind_virtualsvr_realsvr(virtualbindrs_req *kp_proto.S2ABindListenerWithRSReq) kp_proto.KPErrorCode {
	if kpci == nil {
		return kp_proto.KPErrorCode_E_KPERR_FAILED_INSIDERROR
	}

	kpci.m_monitor.Lock()
	defer kpci.m_monitor.Unlock()

	// 校验proxyid和vrrpid是否相同
	ret := kpci.check_proxyid_and_routerid(*virtualbindrs_req.ProxyId, *virtualbindrs_req.VrrpRouterid)
	if ret != kp_proto.KPErrorCode_E_KPERR_OK {
		return ret
	}

	// 判断该virtualsvr是否存在
	virtual_serv, _ := kpci.find_virtual_serv(virtualbindrs_req.NetProtocolName, *virtualbindrs_req.ListenerPort)
	if virtual_serv == nil {
		g_log.Error("VirtualServ net_proto_name[%s] listen_port[%d] is not exist!", *virtualbindrs_req.NetProtocolName,
			*virtualbindrs_req.ListenerPort)
		return kp_proto.KPErrorCode_E_KPERR_FAILED_BINDRS_VIRTUALSVRNOTEXISTS
	}

	// 现有监听器下的rs放入map中
	rs_map := make(map[string]*KPRealServer)
	for index, _ := range virtual_serv.RealSvrInfoLst {
		rs_svr := virtual_serv.RealSvrInfoLst[index]
		rs_map[rs_svr.RsIP] = rs_svr
	}

	for index, _ := range virtualbindrs_req.RealserverLst {
		bindrs_ip := *virtualbindrs_req.RealserverLst[index].Ip
		bindrs_weight := *virtualbindrs_req.RealserverLst[index].Weight
		bindrs_port := *virtualbindrs_req.ListenerPort

		// 判断是否已经存在
		_, ok := rs_map[bindrs_ip]
		if ok {
			// 存在，直接更新
			rs_map[bindrs_ip].RsPort = bindrs_port
			rs_map[bindrs_ip].Weight = bindrs_weight
		} else {
			// 添加
			real_serv := new(KPRealServer)
			real_serv.RsIP = bindrs_ip
			real_serv.Weight = bindrs_weight
			real_serv.RsPort = bindrs_port
			virtual_serv.RealSvrInfoLst = append(virtual_serv.RealSvrInfoLst, real_serv)
		}
	}

	kpci.print_kpci()
	return kp_proto.KPErrorCode_E_KPERR_OK
}

// 监听器解绑rs
func (kpci *KPConfInfo) unbind_virtualsvr_realsvr(virtualandreal_unbindreq *kp_proto.S2AUnBindListenerWithRsReq) kp_proto.KPErrorCode {
	if kpci == nil {
		return kp_proto.KPErrorCode_E_KPERR_FAILED_INSIDERROR
	}

	kpci.m_monitor.Lock()
	defer kpci.m_monitor.Unlock()

	// 校验proxyid和vrrpid是否相同
	ret := kpci.check_proxyid_and_routerid(*virtualandreal_unbindreq.ProxyId, *virtualandreal_unbindreq.VrrpRouterid)
	if ret != kp_proto.KPErrorCode_E_KPERR_OK {
		return ret
	}

	// 判断该virtualsvr是否存在
	virtual_serv, _ := kpci.find_virtual_serv(virtualandreal_unbindreq.NetProtocolName, *virtualandreal_unbindreq.ListenerPort)
	if virtual_serv == nil {
		g_log.Error("VirtualServ net_proto_name[%s] listen_port[%d] is not exist!", *virtualandreal_unbindreq.NetProtocolName,
			*virtualandreal_unbindreq.ListenerPort)
		return kp_proto.KPErrorCode_E_KPERR_FAILED_UNBINDRS_VIRTUALSVRNOTEXISTS
	}

	// 现有监听器下的rs放入map中
	rs_map := make(map[string]*KPRealServer)
	for index, _ := range virtual_serv.RealSvrInfoLst {
		rs_svr := virtual_serv.RealSvrInfoLst[index]
		rs_map[rs_svr.RsIP] = rs_svr
	}

	// 更新rs_map
	for index, _ := range virtualandreal_unbindreq.RealserverLst {
		unbind_rs_ip := virtualandreal_unbindreq.RealserverLst[index]
		_, ok := rs_map[unbind_rs_ip]
		if ok {
			// 从map中排除要删除的rs
			delete(rs_map, unbind_rs_ip)
		}
	}

	// 重新生成rs列表
	virtual_serv.RealSvrInfoLst = virtual_serv.RealSvrInfoLst[:0]
	// 插入
	for key, _ := range rs_map {
		rs_svr := rs_map[key]
		virtual_serv.RealSvrInfoLst = append(virtual_serv.RealSvrInfoLst, rs_svr)
	}

	kpci.print_kpci()
	return kp_proto.KPErrorCode_E_KPERR_OK
}

// 修改监听器
func (kpci *KPConfInfo) update_virtualsvr(updatevirtualsvr_req *kp_proto.S2AModListenerReq) kp_proto.KPErrorCode {
	if kpci == nil {
		return kp_proto.KPErrorCode_E_KPERR_FAILED_INSIDERROR
	}

	kpci.m_monitor.Lock()
	defer kpci.m_monitor.Unlock()

	// 判断该virtualsvr是否存在
	virtual_serv, _ := kpci.find_virtual_serv(updatevirtualsvr_req.NetProtocolName, *updatevirtualsvr_req.ListenerPort)
	if virtual_serv == nil {
		g_log.Error("VirtualServ net_proto_name[%s] listen_port[%d] is not exist!", *updatevirtualsvr_req.NetProtocolName,
			*updatevirtualsvr_req.ListenerPort)
		return kp_proto.KPErrorCode_E_KPERR_FAILED_MODLISTENER_VIRTUALSVRNOTEXISTS
	}

	// 修改数据
	virtual_serv.NetProtoName = *updatevirtualsvr_req.NetProtocolName
	virtual_serv.Port = *updatevirtualsvr_req.ListenerPort
	virtual_serv.Scheduler = *updatevirtualsvr_req.Scheduler
	virtual_serv.DelayLoop = *updatevirtualsvr_req.DelayLoop

	kpci.print_kpci()
	return kp_proto.KPErrorCode_E_KPERR_OK
}

// 执行keepalive reload
func (kpci *KPConfInfo) reload_keepalived() (kp_proto.KPErrorCode, string) {
	if kpci == nil {
		return kp_proto.KPErrorCode_E_KPERR_FAILED_INSIDERROR, ""
	}
	// 生成配置文件
	var err_code kp_proto.KPErrorCode = kp_proto.KPErrorCode_E_KPERR_OK
	var err_reason string = ""

	err_code, err_reason = generate_kpconfigfile()
	if err_code == kp_proto.KPErrorCode_E_KPERR_OK {
		// 执行reload命令
		err_code, err_reason = exec_kpreload()
	}
	return err_code, err_reason
}

// 根据协议和端口找到对应的virtual server，否则返回nil
func (kpci *KPConfInfo) find_virtual_serv(net_proto_name *string, net_port int32) (*KPVritualServer, int) {
	for index, _ := range kpci.VirtualSvrInfoLst {
		virtual_serv := kpci.VirtualSvrInfoLst[index]
		if strings.Compare(virtual_serv.NetProtoName, *net_proto_name) == 0 &&
			virtual_serv.Port == net_port {
			return virtual_serv, index
		}
	}
	return nil, -1
}

// 检测proxyid和vrrp routerid是否相同
func (kpci *KPConfInfo) check_proxyid_and_routerid(proxyid string, vrrp_routerid int32) kp_proto.KPErrorCode {

	if !strings.EqualFold(proxyid, kpci.m_proxyid) {
		g_log.Error("proxyid[%s] not equal KPConfInfo.m_proxyid[%s]", proxyid, g_kpconfinfo.m_proxyid)
		return kp_proto.KPErrorCode_E_KPERR_FAILED_PROXYID_INVALID
	}

	if vrrp_routerid != kpci.VrrpRouterID {
		g_log.Error("vrrp_routerid[%ds] not equal KPConfInfo.VrrpRouterID[%d]", vrrp_routerid, kpci.VrrpRouterID)
		return kp_proto.KPErrorCode_E_KPERR_FAILED_VRRPROUTERID_INVALID
	}

	return kp_proto.KPErrorCode_E_KPERR_OK
}

// 根据协议和port找到对应的listenerid
func (kpci *KPConfInfo) find_listenerid(rsport *int32, protoname *string) (int, string) {
	if rsport != nil && protoname != nil {
		g_log.Debug("RsPort[%d] ProtoName[%s]", *rsport, *protoname)
		kpci.m_monitor.Lock()
		defer kpci.m_monitor.Unlock()

		virtual_serv, _ := kpci.find_virtual_serv(protoname, *rsport)
		if virtual_serv != nil {
			g_log.Debug("virtual_serv.ListenerID[%s]", virtual_serv.ListenerID)
			return 0, virtual_serv.ListenerID
		}
	}
	return -1, ""
}
