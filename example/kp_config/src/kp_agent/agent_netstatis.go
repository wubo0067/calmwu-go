package main

import (
	"fmt"
	"kp_proto"
	"os"
	"strconv"
	"strings"
)

// 网络参数统计
// 每一分钟采集一次，网卡/sys/class/net/ifname/statistics/*
// 网络协议数据 /proc/net/snmp

var (
	netstatis_channel        = make(chan kp_proto.NetStatisticsInfo, 8)
	prev_eth_rxbytes  uint64 = 0
	prev_eth_txbytes  uint64 = 0
)

const (
	eth_rxpcks_filepath_template   = "/sys/class/net/%s/statistics/rx_packets"
	eth_txpcks_filepath_template   = "/sys/class/net/%s/statistics/tx_packets"
	eth_droppcks_filepath_template = "/sys/class/net/%s/statistics/tx_dropped"
	eth_rxbytes_filepath_template  = "/sys/class/net/%s/statistics/rx_bytes"
	eth_txbytes_filepath_template  = "/sys/class/net/%s/statistics/tx_bytes"
	netproto_snmp_filepath         = "/proc/net/snmp"
	eth_data_buff_size             = 64
)

func gather_ethdata(filename string) uint64 {
	hfile, err := os.Open(filename)
	if err != nil {
		g_log.Error("open [%s] failed, reason[%s]", filename, err.Error())
		return 0
	}
	defer hfile.Close()

	eth_data_buff := make([]byte, eth_data_buff_size)
	_, err = hfile.Seek(0, 0)
	if err != nil {
		g_log.Error("seek [%s] failed, reason[%s]", filename, err.Error())
		return 0
	}

	rn, err := hfile.Read(eth_data_buff)
	if err != nil {
		g_log.Error("read [%s] failed, reason[%s]", filename, err.Error())
		return 0
	}

	v := string(eth_data_buff[0 : rn-1])
	n, _ := strconv.ParseUint(v, 10, 64)
	return n
}

func gather_netproto_data(filename string) []uint64 {
	netproto_data := make([]uint64, 6)
	k := 0
	lines, err := readlines_offset_linecount(netproto_snmp_filepath, 0, -1)
	if err != nil {
		g_log.Error("read [%s] failed, reason[%s]", netproto_snmp_filepath, err.Error())
		return netproto_data
	}

	line_count := len(lines)
	for i := 0; i < line_count; i++ {
		line := lines[i]
		r := strings.IndexRune(line, ':')
		if r == -1 {
			g_log.Error("%s is not formatted correctly, expected ':'.", filename)
			return netproto_data
		}

		proto := strings.ToLower(line[:r])
		//g_log.Debug("line[%d] proto[%s]", i, proto)
		if strings.Compare(proto, "tcp") != 0 && strings.Compare(proto, "udp") != 0 {
			// skip protocol and data line
			i++
			continue
		}

		stat_names := strings.Split(line[r+2:], " ")
		//g_log.Debug("stat_names:%v", stat_names)
		// next line
		i++
		stat_values := strings.Split(lines[i][r+2:], " ")
		//g_log.Debug("stat_values:%v", stat_values)
		if len(stat_names) != len(stat_values) {
			g_log.Error("%s is not fomatted correctly, expected same number of columns.", filename)
			return netproto_data
		}

		for j, name := range stat_names {
			if strings.Compare(name, "PassiveOpens") == 0 ||
				strings.Compare(name, "CurrEstab") == 0 ||
				strings.Compare(name, "InSegs") == 0 ||
				strings.Compare(name, "OutSegs") == 0 ||
				strings.Compare(name, "InDatagrams") == 0 ||
				strings.Compare(name, "OutDatagrams") == 0 {
				netproto_data[k], _ = strconv.ParseUint(stat_values[j], 10, 64)
				//g_log.Debug("j[%d] k[%d] %s:%s", j, k, name, stat_values[j])
				k++
			}
		}
	}
	return netproto_data
}

func do_gather_netinfo(ifname string) {
	g_log.Debug("start gather network info!")

	if len(ifname) == 0 {
		g_log.Error("gather network device is invalid!")
		return
	}

	var netstatisinfo kp_proto.NetStatisticsInfo

	eth_rxpcks := gather_ethdata(fmt.Sprintf(eth_rxpcks_filepath_template, ifname))
	eth_txpcks := gather_ethdata(fmt.Sprintf(eth_txpcks_filepath_template, ifname))
	eth_droppcks := gather_ethdata(fmt.Sprintf(eth_droppcks_filepath_template, ifname))
	eth_rxbytes := gather_ethdata(fmt.Sprintf(eth_rxbytes_filepath_template, ifname))
	eth_txbytes := gather_ethdata(fmt.Sprintf(eth_txbytes_filepath_template, ifname))

	netproto_data := gather_netproto_data(netproto_snmp_filepath)

	netstatisinfo.Rxpcks = &eth_rxpcks
	netstatisinfo.Txpcks = &eth_txpcks
	netstatisinfo.Droppcks = &eth_droppcks
	netstatisinfo.Rxbytes = &eth_rxbytes
	netstatisinfo.Txbytes = &eth_txbytes
	netstatisinfo.TcpPassiveopens = &netproto_data[0]
	netstatisinfo.TcpCurrestab = &netproto_data[1]
	netstatisinfo.TcpInsegs = &netproto_data[2]
	netstatisinfo.TcpOutsegs = &netproto_data[3]
	netstatisinfo.UdpIndatarams = &netproto_data[4]
	netstatisinfo.UdpOutdatarams = &netproto_data[5]

	var eth_rxbytes_speed uint32 = 0
	var eth_txbytes_speed uint32 = 0

	if prev_eth_rxbytes != 0 {
		eth_rxbytes_speed = uint32((eth_rxbytes - prev_eth_rxbytes) / 60)
	}

	if prev_eth_txbytes != 0 {
		eth_txbytes_speed = uint32((eth_txbytes - prev_eth_txbytes) / 60)
	}

	prev_eth_rxbytes = eth_rxbytes
	prev_eth_txbytes = eth_txbytes

	netstatisinfo.RxbytesS = &eth_rxbytes_speed
	netstatisinfo.TxbytesS = &eth_txbytes_speed

	g_log.Debug("netdevice [%s] gather info [%s]", ifname, netstatisinfo.String())

	// 发送
	netstatis_channel <- netstatisinfo
	return
}
