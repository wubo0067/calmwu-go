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
	cryptoRand "crypto/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
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

func GenerateRandomPrivateMacAddr() (string, error) {
	buf := make([]byte, 6)
	_, err := cryptoRand.Read(buf)
	if err != nil {
		return "", err
	}

	// Set the local bit for local addresses
	// Addresses in this range are local mac addresses:
	// x2-xx-xx-xx-xx-xx , x6-xx-xx-xx-xx-xx , xA-xx-xx-xx-xx-xx , xE-xx-xx-xx-xx-xx
	buf[0] = (buf[0] | 2) & 0xfe

	hardAddr := net.HardwareAddr(buf)
	return hardAddr.String(), nil
}
