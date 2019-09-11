// +build linux

/*
 * @Author: calmwu
 * @Date: 2017-09-18 10:37:19
 * @Last Modified by: calmwu
 * @Last Modified time: 2019-03-14 20:58:00
 * @Comment:
 */

package utils

import (
	"bufio"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
)

type NetErrorType int

const (
	NETERR_TYPE_NO                  NetErrorType = iota //
	NETERR_TYPE_DNSERROR                                // *net.DNSError
	NETERR_TYPE_INVALIDADDERROR                         // *net.InvalidAddrError
	NETERR_TYPE_UNKNOWNNETWORKERROR                     // *net.UnknownNetworkError
	NETERR_TYPE_ADDERROR                                // *net.AddrError
	NETERR_TYPE_DNSCONFIGERROR                          // *net.DNSConfigError
	NETERR_TYPE_OS_SYSCALLERROR                         // *os.SyscallError--->syscall.Errno syscall.ECONNREFUSED syscall.ETIMEDOUT
)

var reusePort = 0x0F

func GetIPByIfname(ifname string) (string, error) {
	local_ip := string("UnknownIP")
	iface_lst, err := net.Interfaces()
	if err == nil {
		for _, iface := range iface_lst {
			if iface.Name == ifname {
				//得到地址
				local_addrs, _ := iface.Addrs()
				local_ip = local_addrs[0].String()

				temp := strings.Split(local_ip, "/")
				return temp[0], nil
			}
		}
	}
	return "", err
}

func SetRecvBuf(c *net.TCPConn, recvBufSize int) error {
	size := recvBufSize
	var err error
	for size > 0 {
		if err = c.SetReadBuffer(size); err == nil {
			return nil
		}
		size = size / 2
	}
	return err
}

func SetKeepAlive(fd, secs int) error {
	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1); err != nil {
		return err
	}
	if err := syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPINTVL, secs); err != nil {
		return err
	}
	return syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPIDLE, secs)
}

func SetReuseAddrAndPort(socketFD int) error {
	var err error
	if err = syscall.SetsockoptInt(socketFD, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		return err
	}

	if err = syscall.SetsockoptInt(socketFD, syscall.SOL_SOCKET, reusePort, 1); err != nil {
		return err
	}
	return nil
}

func MaxListenerBacklog() int {
	fd, err := os.Open("/proc/sys/net/core/somaxconn")
	if err != nil {
		return syscall.SOMAXCONN
	}
	defer fd.Close()

	rd := bufio.NewReader(fd)
	line, err := rd.ReadString('\n')
	if err != nil {
		return syscall.SOMAXCONN
	}

	f := strings.Fields(line)
	if len(f) < 1 {
		return syscall.SOMAXCONN
	}

	n, err := strconv.Atoi(f[0])
	if err != nil || n == 0 {
		return syscall.SOMAXCONN
	}

	// Linux stores the backlog in a uint16.
	// Truncate number to avoid wrapping.
	// See issue 5030.
	if n > 1<<16-1 {
		n = 1<<16 - 1
	}

	return n
}

func SockaddrToAddr(sa syscall.Sockaddr) net.Addr {
	var a net.Addr
	switch sa := sa.(type) {
	case *syscall.SockaddrInet4:
		a = &net.TCPAddr{
			IP:   append([]byte{}, sa.Addr[:]...),
			Port: sa.Port,
		}
	case *syscall.SockaddrInet6:
		var zone string
		if sa.ZoneId != 0 {
			if ifi, err := net.InterfaceByIndex(int(sa.ZoneId)); err == nil {
				zone = ifi.Name
			}
		}
		if zone == "" && sa.ZoneId != 0 {
		}
		a = &net.TCPAddr{
			IP:   append([]byte{}, sa.Addr[:]...),
			Port: sa.Port,
			Zone: zone,
		}
	case *syscall.SockaddrUnix:
		a = &net.UnixAddr{Net: "unix", Name: sa.Name}
	}
	return a
}

// https://liudanking.com/network/go-%E4%B8%AD%E5%A6%82%E4%BD%95%E5%87%86%E7%A1%AE%E5%9C%B0%E5%88%A4%E6%96%AD%E5%92%8C%E8%AF%86%E5%88%AB%E5%90%84%E7%A7%8D%E7%BD%91%E7%BB%9C%E9%94%99%E8%AF%AF/
func NetErrorCheck(err error) (isNetError bool, netErrEnum NetErrorType, netErr interface{}) {
	if netErr, ok := err.(net.Error); ok {
		if opErr, ok := netErr.(*net.OpError); ok {
			switch t := opErr.Err.(type) {
			case *net.DNSError:
				return true, NETERR_TYPE_DNSERROR, t
			case *net.InvalidAddrError:
				return true, NETERR_TYPE_INVALIDADDERROR, t
			case *net.UnknownNetworkError:
				return true, NETERR_TYPE_UNKNOWNNETWORKERROR, t
			case *net.AddrError:
				return true, NETERR_TYPE_ADDERROR, t
			case *net.DNSConfigError:
				return true, NETERR_TYPE_DNSCONFIGERROR, t
			case *os.SyscallError:
				if sysErr, ok := t.Err.(syscall.Errno); ok {
					// https://golang.org/pkg/syscall/#Errno
					return true, NETERR_TYPE_OS_SYSCALLERROR, sysErr
				}
			}
		}
	}
	return false, NETERR_TYPE_NO, nil
}
