// +build linux

/*
 * @Author: calmwu
 * @Date: 2017-12-01 14:41:03
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-12-08 16:18:15
 * @Comment: 网络管理
 */

package transport

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wubo0067/calmwu-go/utils"
)

const (
	DEF_TCPRECVBUF_SIZE int = 1 << 20
)

type NetTransportConfig struct {
	sessionReadTimeout      time.Duration
	sessionIdleTimeout      time.Duration
	sesssionBusinssChanSize int
	maxSessionCount         int32
	netPkgMaxSize           int
}

type NetTransport struct {
	config           *NetTransportConfig
	tcpListener      *net.TCPListener
	shutdown         int32
	shutdownCh       chan struct{}
	connCh           chan *net.TCPConn
	wg               sync.WaitGroup
	toBussinessCh    chan *NetSessionData // from transport ==to==> bussiness module chan
	frombussinessCh  chan *NetSessionData // from business module ==to==> transform chan
	sessionIDGen     uint32
	currSessionCount int32

	sessionMapGuard *sync.RWMutex
	sessionMap      map[uint32]*NetSession
}

func NewDefaultNetTransportConfig() *NetTransportConfig {
	return &NetTransportConfig{
		sessionIdleTimeout:      10 * time.Second,
		sessionReadTimeout:      5 * time.Millisecond,
		sesssionBusinssChanSize: 10240,
		maxSessionCount:         10240,
		netPkgMaxSize:           1048576,
	}
}

func StartNetTransport(listenIP string, listenPort int, config *NetTransportConfig) (*NetTransport, error) {
	ip := net.ParseIP(listenIP)

	tcpAddr := &net.TCPAddr{IP: ip, Port: listenPort}
	// 启动监听
	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		utils.ZLog.Errorf("Failed to start TCP listener on %q port %d. reason[%s]",
			listenIP, listenPort, err.Error())
		return nil, err
	}

	utils.ZLog.Debug("NetTransport confiig:%+v", *config)

	netTransport := new(NetTransport)
	netTransport.tcpListener = tcpListener
	netTransport.shutdownCh = make(chan struct{})
	netTransport.connCh = make(chan *net.TCPConn)
	netTransport.toBussinessCh = make(chan *NetSessionData, config.sesssionBusinssChanSize)
	netTransport.frombussinessCh = make(chan *NetSessionData, config.sesssionBusinssChanSize)
	netTransport.config = config
	netTransport.sessionMapGuard = new(sync.RWMutex)
	netTransport.sessionMap = make(map[uint32]*NetSession)
	netTransport.wg.Add(2)

	go netTransport.tcpListenerRoutine()
	go netTransport.tcpNewConnectRoutine()

	utils.ZLog.Info("NetTransport started! listen on[%s:%d]", listenIP, listenPort)
	return netTransport, nil
}

func (nt *NetTransport) ShutDown() {
	defer utils.ZLog.Info("NetTransport ShutDown!!")
	atomic.StoreInt32(&nt.shutdown, 1)
	close(nt.shutdownCh)

	if nt.tcpListener != nil {
		nt.tcpListener.Close()
	}

	nt.wg.Wait()
}

func (nt *NetTransport) tcpListenerRoutine() {
	defer utils.ZLog.Info("tcpListenerRoutine exit!")
	defer nt.wg.Done()

	for {
		conn, err := nt.tcpListener.AcceptTCP()
		if err != nil {
			if s := atomic.LoadInt32(&nt.shutdown); s == 1 {
				utils.ZLog.Warn("tcpListenerRoutine shutdown flag is set!")
				break
			}

			utils.ZLog.Errorf("Error accepting Tcp connection! reason[%s]", err.Error())
			continue
		}

		if nt.CurrSessionCount() < nt.config.maxSessionCount {
			nt.connCh <- conn
			utils.ZLog.Debug("Accept new connect[%s]", conn.RemoteAddr().String())
		} else {
			conn.Close()
			utils.ZLog.Errorf("The session count exceeds the limit!")
		}
	}
}

func (nt *NetTransport) tcpNewConnectRoutine() {
	defer utils.ZLog.Info("tcpNewConnectRoutine exit!")
	defer nt.wg.Done()

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case newConn := <-nt.connCh:

			if nt.CurrSessionCount() < nt.config.maxSessionCount {
				nt.wg.Add(1)
				nt.addSession(newConn)
			} else {
				newConn.Close()
				utils.ZLog.Errorf("The session count exceeds the limit!")
			}

		case sessionData, ok := <-nt.frombussinessCh:
			if ok {
				// 将数据转发给session
				nt.dispatchSessionData(sessionData)
			}
		case <-ticker.C:
			utils.ZLog.Debug("Current Session Count:%d", nt.CurrSessionCount())
		case <-nt.shutdownCh:
			utils.ZLog.Warn("tcpNewConnectRoutine receive shutdown notify")
			return
		}
	}
}

func (nt *NetTransport) ReadDataCh() <-chan *NetSessionData {
	return nt.toBussinessCh
}

func (nt *NetTransport) WriteData(data *NetSessionData) {
	nt.frombussinessCh <- data
}

func (nt *NetTransport) CurrSessionCount() int32 {
	return atomic.LoadInt32(&(nt.currSessionCount))
}

func (nt *NetTransport) addSession(newConn *net.TCPConn) {
	nt.sessionMapGuard.Lock()
	defer nt.sessionMapGuard.Unlock()
	nt.sessionIDGen++
	newSession := NewSession(nt.sessionIDGen, newConn, nt)
	nt.sessionMap[nt.sessionIDGen] = newSession
	atomic.AddInt32(&(nt.currSessionCount), 1)
}

func (nt *NetTransport) removeSession(sessionID uint32) {
	nt.sessionMapGuard.Lock()
	defer nt.sessionMapGuard.Unlock()

	if _, ok := nt.sessionMap[sessionID]; ok {
		delete(nt.sessionMap, sessionID)
		atomic.AddInt32(&(nt.currSessionCount), -1)
	} else {
		utils.ZLog.Errorf("sessionID[%d] is not exist!", sessionID)
	}
}

func (nt *NetTransport) dispatchSessionData(sessionData *NetSessionData) {
	nt.sessionMapGuard.RLock()
	defer nt.sessionMapGuard.RUnlock()
	if session, ok := nt.sessionMap[sessionData.SessionID]; ok {
		session.frombussinessCh <- sessionData
	} else {
		utils.ZLog.Errorf("Session[%d] can not find in sessionMap!", sessionData.SessionID)
	}
}
