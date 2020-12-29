/*
 * @Author: calmwu
 * @Date: 2017-09-18 10:37:19
 * @Last Modified by:   calmwu
 * @Last Modified time: 2017-09-18 10:37:19
 * @Comment:
 */

package base

import (
	"net"
	"strings"
)

func GetIPByIfname(ifname string) (*string, error) {
	local_ip := string("UnknownIP")
	iface_lst, err := net.Interfaces()
	if err == nil {
		for _, iface := range iface_lst {
			if iface.Name == ifname {
				//得到地址
				local_addrs, _ := iface.Addrs()
				local_ip = local_addrs[0].String()

				temp := strings.Split(local_ip, "/")
				return &temp[0], nil
			}
		}
	}
	return nil, err
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
