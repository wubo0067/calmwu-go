/*
 * @Author: calmwu
 * @Date: 2017-12-01 16:01:45
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-12-08 16:05:17
 * @Comment:
 */

package transport

import (
	"bufio"
	"bytes"
	"crypto/cipher"
	"encoding/hex"
	"io"
	"net"
	"runtime"
	"calmwu-go/utils"
	"sailcraft/network/protocol"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/golang/protobuf/proto"
)

type SessionState int
type SessionPeekState int

const (
	E_SESSIONSTATE_VERIFIY_SYN SessionState = iota
	E_SESSIONSTATE_VERIFIY_ACK
	E_SESSIONSTATE_CONNECTED
)

const (
	E_SESSIONPEEKSTATE_HEAD SessionPeekState = iota
	E_SESSIONPEEKSTATE_WHOLE
)

const (
	MAX_SELECTTIMES      int    = 10
	CLOSESELF_CMD        int16  = ^int16(0) - 1
	CLOSESELF_MAGICVALUE uint32 = ^uint32(0) - 1016
	CLOSESELF_MSGID      uint32 = ^uint32(0) - 78117
	CLOSESLEF_DATALEN    int32  = int32(^uint32(0)>>1) - 120703
)

type NetSession struct {
	sessionID uint32
	conn      *net.TCPConn
	reader    *bufio.Reader

	frombussinessCh chan *NetSessionData // from business module ==to==> session chan
	nt              *NetTransport
	sessionState    SessionState
	encryptionKey   cipher.Block
	verifyBuf       []byte
	totalReadBytes  int
	shutdown        int32
}

func NewSession(sessionID uint32, conn *net.TCPConn, nt *NetTransport) *NetSession {
	newSession := &NetSession{
		sessionID:       sessionID,
		conn:            conn,
		reader:          bufio.NewReader(conn),
		frombussinessCh: make(chan *NetSessionData, nt.config.sesssionBusinssChanSize),
		nt:              nt,
		sessionState:    E_SESSIONSTATE_VERIFIY_SYN,
	}

	utils.ZLog.Debug("Create new session for %s", newSession.RemoteAddr())

	//notifyCh := newSession.sessionIdleMonitor()
	go newSession.handleBusiness()
	go newSession.handleConn()
	return newSession
}

func (ns *NetSession) handleConn() {
	utils.ZLog.Info("Session[%d] New incoming connection[%s]", ns.sessionID, ns.conn.RemoteAddr().String())

	defer func() {
		ns.conn.Close()
		ns.nt.wg.Done()
		ns.nt.removeSession(ns.sessionID)
		ns.notfityBusinessConnStop()

		if err := recover(); err != nil {
			// 回收panic，防止整个服务panic
			utils.ZLog.Errorf("Session[%d] painc! reson:%v, stack:%s", ns.sessionID, err, utils.GetCallStack())
		}
		utils.ZLog.Debug("Session[%d] disconnect with client[%s]", ns.sessionID, ns.RemoteAddr())
	}()

	//sessionReadNow := time.Now()
	var peekSize int = protocol.ProtoC2SHeadSize
	var peekState SessionPeekState = E_SESSIONPEEKSTATE_HEAD
	var protoHead *protocol.ProtoC2SHead
	var err error
	var data []byte
	readDataTime := time.Now()

	// 这是默认行为
	//ns.conn.SetNoDelay(true)
	utils.SetRecvBuf(ns.conn, DEF_TCPRECVBUF_SIZE)

	for {
		if s := atomic.LoadInt32(&ns.shutdown); s == 1 {
			utils.ZLog.Warn("Session[%d] Shutdown flag is set!", ns.sessionID)
			return
		}
		// 读取头
		ns.conn.SetDeadline(time.Now().Add(time.Second))
		data, err = ns.reader.Peek(peekSize)

		if err != nil {
			// 错误处理
			if err == io.EOF {
				utils.ZLog.Warn("Session[%d] client[%s] disconnect the session. reason[%s]",
					ns.sessionID, ns.RemoteAddr(), err.Error())
				return
			} else if oerr, ok := err.(*net.OpError); ok {
				if oerr.Timeout() && time.Since(readDataTime) >= ns.nt.config.sessionIdleTimeout {
					// read timeout
					utils.ZLog.Warn("Session[%d] idle exceed limit readDataTime[%s] now[%s], so will disconnect!",
						ns.sessionID, utils.TimeName(readDataTime), utils.GetTimeStampSec())
					return
				} else if oerr.Temporary() {
					// EAGAIN
					runtime.Gosched()
					continue
				} else if oerr.Err == syscall.ECONNRESET {
					// read RST
					utils.ZLog.Warn("Session[%d] receive RESET, so will disconnect! SourceAddr[%s] Addr[%s] reason[%s]",
						ns.sessionID, oerr.Source.String(), oerr.Addr.String(), err.Error())
					return
				}
			} else {
				// peekSize > reader的缓冲区，会返回错误ErrBufferFull
				utils.ZLog.Errorf("Session[%d] err:%s buffSize:%d", ns.sessionID, err.Error(), ns.reader.Buffered())
				return
			}
		} else {
			//数据处理
			readDataTime = time.Now()
			switch peekState {
			case E_SESSIONPEEKSTATE_HEAD:
				protoHead, err = protocol.UnPackProtoHead(data)
				if err != nil {
					utils.ZLog.Errorf("Session[%d] UnPackProtoHead failed so will disconnect. reason[%s]", ns.sessionID, err.Error())
					return
				} else {
					// 对头进行校验
					if protoHead.MagicValue != protocol.ProtoC2SMagic {
						utils.ZLog.Errorf("Session[%d] check head magic failed so will disconnect", ns.sessionID)
						return
					}

					// 判断包大小的合法性
					if protoHead.DataLen < 0 || int(protoHead.DataLen) > ns.nt.config.netPkgMaxSize {
						utils.ZLog.Errorf("Session[%d] check head datalen[%d] is illegal", ns.sessionID, protoHead.DataLen)
						return
					}

					//utils.ZLog.Debug("head:%+v", *protoHead)
					peekSize = protocol.ProtoC2SHeadSize + int(protoHead.DataLen)
					peekState = E_SESSIONPEEKSTATE_WHOLE
					//utils.ZLog.Debug("Session[%d] package whole size[%d] protoHead:%+v", ns.sessionID, peekSize, *protoHead)
				}
			case E_SESSIONPEEKSTATE_WHOLE:
				if ns.handleNetPkg(protoHead, data[protocol.ProtoC2SHeadSize:]) < 0 {
					utils.ZLog.Errorf("Session[%d] handleNetPkg failed so will disconnect", ns.sessionID)
					return
				}
				ns.reader.Discard(peekSize)
				peekSize = protocol.ProtoC2SHeadSize
				peekState = E_SESSIONPEEKSTATE_HEAD
			}
		}
	}
}

func (ns *NetSession) handleBusiness() {

	for {
		select {
		case sessionData := <-ns.frombussinessCh:
			switch sessionData.Cmd {
			case E_SESSIONCMD_STOPCONN:
				// bussiness主动断开连接
				if sessionData.SessionID == ns.sessionID && ns.sessionState == E_SESSIONSTATE_CONNECTED {
					utils.ZLog.Info("Session[%d] recv E_SESSIONCMD_STOPCONN from business, now will disconnect session!",
						ns.sessionID)
				} else {
					utils.ZLog.Info("Session[%d] recv E_SESSIONCMD_STOPCONN, but sessionid[%d] is mismatching",
						ns.sessionID)
				}
			case E_SESSIONCMD_TRANSFER:
				if sessionData.SessionID == ns.sessionID && ns.sessionState == E_SESSIONSTATE_CONNECTED {
					//
					transferData, err := protocol.FlodPayloadData(sessionData.Data, ns.encryptionKey)
					if err != nil {
						utils.ZLog.Errorf("Session[%d] transfer data FlodPayloadData failed! reason[%s]",
							ns.sessionID, err.Error())
					} else {
						payload, err := protocol.PackTransferData(sessionData.MsgId, transferData)
						if err != nil {
							utils.ZLog.Errorf("Session[%d] transfer data FlodPayloadData failed! reason[%s]",
								ns.sessionID, err.Error())
						} else {
							ns.conn.Write(payload)
						}
					}
				}
			}
		case <-ns.nt.shutdownCh:
			utils.ZLog.Info("Session[%d] handleBusiness receive shutdown notify!", ns.sessionID)
			atomic.StoreInt32(&ns.shutdown, 1)
			return
		}
	}
	return
}

func (ns NetSession) RemoteAddr() string {
	return ns.conn.RemoteAddr().String()
}

func (ns *NetSession) handleNetPkg(head *protocol.ProtoC2SHead, data []byte) int {
	switch ns.sessionState {
	case E_SESSIONSTATE_VERIFIY_SYN:
		// 处理E_C2S_CMD_SYN包
		if head.Cmd == protocol.E_C2S_CMD_SYN {
			var syn protocol.ProtoSyn
			err := proto.Unmarshal(data, &syn)
			if err != nil {
				utils.ZLog.Errorf("Session[%d] unmarshal ProtoSyn failed! reason[%s]",
					ns.sessionID, err.Error())
				return -1
			}

			utils.ZLog.Debug("Session[%d] synPkg:%+v", ns.sessionID, syn)

			dhKey, err := utils.GenerateDHKey()
			if err != nil {
				utils.ZLog.Errorf("Session[%d] GenerateDHKey failed! reason[%s]",
					ns.sessionID, err.Error())
				return -1
			}

			// 根据对方发过来的ka，计算自己的秘钥串
			key, err := utils.GenerateEncryptionKey(syn.DHClientPubKey, dhKey)
			if err != nil {
				utils.ZLog.Errorf("Session[%d] GenerateEncryptionKey failed! reason[%s]",
					ns.sessionID, err.Error())
				return -1
			}

			secretKeyB := hex.EncodeToString(key)
			utils.ZLog.Debug("golang server secretKeyB[%s]", secretKeyB)

			// aes-256
			ns.encryptionKey, err = utils.NewCipherBlock(key[:32])
			if err != nil {
				utils.ZLog.Errorf("Session[%d] NewCipherBlock failed! reason[%s]",
					ns.sessionID, err.Error())
				return -1
			}

			ns.verifyBuf = syn.VerifyBuf

			var asyn protocol.ProtoAsyn
			asyn.DHServerPubKey = dhKey.Bytes()
			payload, err := protocol.PackAsynMsg(head.MsgId, &asyn)
			if err != nil {
				utils.ZLog.Errorf("Session[%d] PackAsynMsg failed! reason[%s]",
					ns.sessionID, err.Error())
				return -1
			}

			// 发送出去
			ns.conn.Write(payload)

			// 交换密钥
			ns.sessionState = E_SESSIONSTATE_VERIFIY_ACK
		} else {
			utils.ZLog.Errorf("Session[%d] currState[%s] receive Cmd[%s] is invalid!", ns.sessionID,
				ns.sessionState.String(), head.Cmd.String())
			return -1
		}
	case E_SESSIONSTATE_VERIFIY_ACK:
		// 处理E_C2S_CMD_ACK包
		if head.Cmd == protocol.E_C2S_CMD_ACK {
			var ack protocol.ProtoAck
			err := proto.Unmarshal(data, &ack)
			if err != nil {
				utils.ZLog.Errorf("Session[%d] Unmarshal ProtoAck failed! reason[%s]", ns.sessionID, err.Error())
				return -1
			}
			// 对校验数据进行加密，比较
			myCipherText, err := utils.EncryptPlainText(ns.encryptionKey, ns.verifyBuf)
			if err != nil {
				utils.ZLog.Errorf("Session[%d] EncryptPlainText failed! reason[%s]", ns.sessionID,
					err.Error())
				return -1
			}
			if bytes.Equal(ack.CipherText, myCipherText) {
				// 握手成功
				utils.ZLog.Info("Session[%d] remote[%s] handshake ok!", ns.sessionID, ns.RemoteAddr())
				ns.notifyBusinessConnStart()
				ns.sessionState = E_SESSIONSTATE_CONNECTED
			} else {
				utils.ZLog.Errorf("Session[%d] encrypt check failed! CipherText:%v, myCipherText:%v",
					[]byte(ack.CipherText), myCipherText)
				return -1
			}
		} else {
			utils.ZLog.Errorf("Session[%d] currState[%s] receive Cmd[%s] is invalid!", ns.sessionID,
				ns.sessionState.String(), head.Cmd.String())
			return -1
		}
	case E_SESSIONSTATE_CONNECTED:
		// 处理E_TRANSFER
		if head.Cmd == protocol.E_TRANSFER {
			// 将数据转发给bussiness
			payload, err := protocol.UnFlodPayloadData(data, ns.encryptionKey)
			if err != nil {
				utils.ZLog.Errorf("Session[%d] UnFlodPayloadData failed! reason[%s]", ns.sessionID, err.Error())
			} else {
				// 透传数据
				netSessionData := PoolGetSessionData()
				netSessionData.Cmd = E_SESSIONCMD_TRANSFER
				netSessionData.SessionID = ns.sessionID
				netSessionData.MsgId = head.MsgId
				netSessionData.Data = payload
				ns.nt.toBussinessCh <- netSessionData
			}
		} else if head.Cmd == protocol.E_HEARTBEAT {
			var heartbeat protocol.ProtoHeartBeat
			err := proto.Unmarshal(data, &heartbeat)
			if err != nil {
				utils.ZLog.Errorf("Session[%d] Unmarshal ProtoHeartBeat failed! reason[%s]", ns.sessionID, err.Error())
				return -1
			}
			ns.handleHeartBeat(head.MsgId, heartbeat.Timestamp)
		} else {
			utils.ZLog.Errorf("Session[%d] currState[%s] receive Cmd[%s] is invalid!", ns.sessionID,
				ns.sessionState.String(), head.Cmd.String())
			return -1
		}
	}
	return 0
}

func (ns *NetSession) handleHeartBeat(msgId uint32, unix int64) {
	// 回应
	//utils.ZLog.Debug("handle remote[%s] heartbeat", ns.RemoteAddr())
	heartbeat, _ := protocol.PackHeartBeat(msgId, unix)
	ns.conn.Write(heartbeat)
}

func (ns *NetSession) notifyBusinessConnStart() {
	start := PoolGetSessionData()
	start.Cmd = E_SESSIONCMD_STARTCONN
	start.SessionID = ns.sessionID
	ns.nt.toBussinessCh <- start
	utils.ZLog.Debug("Session[%d] send [%s] to bussiness", ns.sessionID, ns.sessionState.String())
	return
}

func (ns *NetSession) notfityBusinessConnStop() {
	if ns.sessionState == E_SESSIONSTATE_CONNECTED {
		stop := PoolGetSessionData()
		stop.Cmd = E_SESSIONCMD_STOPCONN
		stop.SessionID = ns.sessionID
		ns.nt.toBussinessCh <- stop
		utils.ZLog.Debug("Session[%d] send [%s] to bussiness", ns.sessionID, ns.sessionState.String())
	} else {
		utils.ZLog.Warn("Session[%d] currState[%s] no need to send cmd[E_SESSIONSTATE_CONNECTED]",
			ns.sessionID, ns.sessionState.String())
	}
}
