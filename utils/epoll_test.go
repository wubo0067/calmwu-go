/*
 * @Author: calmwu
 * @Date: 2019-02-22 14:57:01
 * @Last Modified by: calmwu
 * @Last Modified time: 2019-02-22 16:37:05
 */

package utils

import (
	"fmt"
	"net"
	"os"
	"testing"
)

func TestGetNetConnSocketFD(t *testing.T) {
	ip := net.ParseIP("127.0.0.1")

	tcpAddr := &net.TCPAddr{IP: ip, Port: 8889}
	// 启动监听
	tcpListener, _ := net.ListenTCP("tcp", tcpAddr)

	listenerFD := TcpListenerSocketFD(tcpListener)
	fmt.Printf("tcpListener:%+v, listenerFD:%d", tcpListener, listenerFD)
}

func TestEpollAdd(t *testing.T) {
	ip := net.ParseIP("127.0.0.1")
	tcpAddr := &net.TCPAddr{IP: ip, Port: 8888}
	tcpListener, _ := net.ListenTCP("tcp", tcpAddr)

	epoll, err := NewEpoll()
	if err != nil {
		t.Error(err.Error())
		return
	}

	_, err = epoll.Add(tcpListener, nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Println("EpollAdd successed!")
}

func TestEpollEcho(t *testing.T) {
	ip := net.ParseIP("0.0.0.0")
	tcpAddr := &net.TCPAddr{IP: ip, Port: 8887}
	tcpListener, _ := net.ListenTCP("tcp", tcpAddr)

	epoll, err := NewEpoll()
	if err != nil {
		t.Error(err.Error())
		return
	}

	_, err = epoll.Add(tcpListener, nil)
	if err != nil {
		t.Error(err.Error())
		return
	}

	for {
		conns, err := epoll.Wait(100)
		if err != nil {
			fmt.Fprintf(os.Stderr, "err:%s\n", err.Error())
			continue
		}

		for _, conn := range conns {
			if conn.ConnType == EPOLLConnTypeTCPLISTENER {
				listener := conn.ConnHolder.(*net.TCPListener)
				client, err := listener.AcceptTCP()
				if err != nil {
					fmt.Fprintf(os.Stderr, "AcceptTCP failed! err:%s\n", err.Error())
					return
				}
				fmt.Printf("accept new connect, %s\n", client.RemoteAddr().String())
				// 将client加入epoll
				epoll.Add(client, nil)
			} else if conn.ConnType == EPOLLConnTypeTCPCONN {
				clientConn := conn.ConnHolder.(*net.TCPConn)

				// 读取
				buffer := make([]byte, 32)
				n, err := clientConn.Read(buffer)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Read failed! err:%s\n", err.Error())
					return
				} else {
					fmt.Printf("Read %d bytes\n", n)
				}

				n, err = clientConn.Write(buffer[:n])
				if err != nil {
					fmt.Fprintf(os.Stderr, "Write failed! err:%s\n", err.Error())
				} else {
					fmt.Printf("Write %d bytes\n", n)
				}
			}
		}
	}
}

func TestMaxListenerBacklog(t *testing.T) {
	listenerBacklog := MaxListenerBacklog()
	fmt.Printf("listenerBacklog:%d\n", listenerBacklog)
}
