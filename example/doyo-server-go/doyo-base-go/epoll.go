// +build linux

/*
 * @Author: calmwu
 * @Date: 2019-02-22 14:38:29
 * @Last Modified by: calmwu
 * @Last Modified time: 2019-02-22 20:06:27
 */

package base

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"sync"
	"syscall"

	"github.com/gorilla/websocket"
	"golang.org/x/sys/unix"
)

type epollConnType int

const (
	EPOLL_CONNTYPE_TCPCONN epollConnType = iota
	EPOLL_CONNTYPE_TCPLISTENER
	EPOLL_CONNTYPE_UDP
	EPOLL_CONNTYPE_WEBSOCKET
)

/*
const (
	POLLIN    = 0x1
	POLLPRI   = 0x2
	POLLOUT   = 0x4
	POLLRDHUP = 0x2000
	POLLERR   = 0x8
	POLLHUP   = 0x10
	POLLNVAL  = 0x20
)
*/
type EpollConn struct {
	ConnHolder    interface{}   // golang各种连接对象
	ConnArg       interface{}   // 附加参数
	connType      epollConnType // 连接类型
	TriggerEvents uint32        // EpollEvent返回的事件类型
}

type epoll struct {
	fd          int
	connections map[int]*EpollConn
	lock        *sync.RWMutex
}

func NewEpoll() (*epoll, error) {
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	return &epoll{
		fd:          fd,
		lock:        &sync.RWMutex{},
		connections: make(map[int]*EpollConn),
	}, nil
}

func (ep *epoll) Add(conn, connArg interface{}) (int, error) {
	//connType := reflect.Indirect(reflect.ValueOf(conn)).Type()
	var socketFD int
	var econn EpollConn

	switch realConn := conn.(type) {
	case *net.TCPConn:
		socketFD = TcpConnSocketFD(realConn)
		econn.connType = EPOLL_CONNTYPE_TCPCONN
	case *net.TCPListener:
		socketFD = TcpListenerSocketFD(realConn)
		econn.connType = EPOLL_CONNTYPE_TCPLISTENER
	case *net.UDPConn:
		socketFD = UdpConnSocketFD(realConn)
		econn.connType = EPOLL_CONNTYPE_UDP
	case *websocket.Conn:
		socketFD = GorillaConnSocketFD(realConn)
		econn.connType = EPOLL_CONNTYPE_WEBSOCKET
	default:
		return -1, errors.New(fmt.Sprintf("conn type:%s is not support\n", reflect.Indirect(reflect.ValueOf(conn)).Type().Name()))
	}

	econn.ConnHolder = conn
	econn.ConnArg = connArg

	err := unix.EpollCtl(ep.fd, syscall.EPOLL_CTL_ADD, socketFD, &unix.EpollEvent{Events: unix.POLLIN | unix.POLLHUP, Fd: int32(socketFD)})
	if err != nil {
		return -1, err
	}
	ep.lock.Lock()
	defer ep.lock.Unlock()
	ep.connections[socketFD] = &econn
	return socketFD, nil
}

func (ep *epoll) Remove(socketFD int) error {
	err := unix.EpollCtl(ep.fd, syscall.EPOLL_CTL_DEL, socketFD, nil)
	if err != nil {
		return err
	}

	ep.lock.Lock()
	defer ep.lock.Unlock()
	delete(ep.connections, socketFD)
	return nil
}

func (ep *epoll) Wait(milliseconds int) ([]*EpollConn, error) {
	events := make([]unix.EpollEvent, 1024)

	var n int
	var err error

	for {
		n, err = unix.EpollWait(ep.fd, events, milliseconds)
		if err != nil {
			if err == syscall.EINTR {
				continue
			} else {
				return nil, err
			}
		} else {
			break
		}
	}

	ep.lock.RLock()
	defer ep.lock.RUnlock()
	var connections []*EpollConn
	for i := 0; i < n; i++ {
		if conn, exist := ep.connections[int(events[i].Fd)]; exist {
			conn.TriggerEvents = events[i].Events
			connections = append(connections, conn)
		}
	}
	// 返回可读的网络连接
	return connections, nil
}

func TcpConnSocketFD(conn *net.TCPConn) int {
	// 就算是私有成员通过反射还是可以获取
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}

func UdpConnSocketFD(conn *net.UDPConn) int {
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}

func TcpListenerSocketFD(listener *net.TCPListener) int {
	fdVal := reflect.Indirect(reflect.ValueOf(listener)).FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}

func GorillaConnSocketFD(conn *websocket.Conn) int {
	// Elem()从返回的interface中获取真实的对象
	connVal := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn").Elem()
	tcpConn := reflect.Indirect(connVal).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}
