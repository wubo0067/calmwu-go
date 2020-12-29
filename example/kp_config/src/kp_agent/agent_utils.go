package main

import (
	//"path"
	"bufio"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"unsafe"

	// 导入日志模块
	l4g "log4go"
)

// 日志句柄
var g_log l4g.Logger = make(l4g.Logger)

func init_log(logname string) int {
	// file_info, err := os.Stat(*cmd_params_logpath)

	// // 判断路径是否存在
	// if err != nil || os.IsNotExist(err) {
	// 	fmt.Fprintf(os.Stderr, "logpath[%s] is not exist!\n", *cmd_params_logpath)
	// 	return -1
	// }

	// // 检查是不是目录
	// if !file_info.IsDir() {
	// 	fmt.Fprintf(os.Stderr, "logpath[%s] is not Directory!\n", *cmd_params_logpath)
	// 	return -1
	// }

	logfile := fmt.Sprintf("/var/log/%s", logname)
	log_writer := l4g.NewFileLogWriter(logfile, false)
	log_writer.SetRotate(true)
	log_writer.SetRotateSize(50 * 1024 * 1024)
	log_writer.SetRotateMaxBackup(10)
	g_log.AddFilter("normal", l4g.FINE, log_writer)
	return 0
}

func stack_trace(all bool) string {
	// Reserve 10K buffer at first
	buf := make([]byte, 10240)

	for {
		size := runtime.Stack(buf, all)
		// The size of the buffer may be not enough to hold the stacktrace,
		// so double the buffer size
		if size == len(buf) {
			buf = make([]byte, len(buf)<<1)
			continue
		}
		break
	}
	return string(buf)
}

func is_littleendian() bool {
	var i int32 = 0x01020304
	u := unsafe.Pointer(&i)
	pb := (*byte)(u)
	b := *pb
	return (b == 0x04)
}

func get_localip_by_ifname(ifname string) *string {
	local_ip := string("UnknownIP")
	iface_lst, err := net.Interfaces()
	if err == nil {
		for _, iface := range iface_lst {
			if iface.Name == ifname {
				//得到地址
				local_addrs, _ := iface.Addrs()
				local_ip = local_addrs[0].String()
			}
		}
	}
	temp := strings.Split(local_ip, "/")
	return &temp[0]
}

func check_file_exist(file_path string) (bool, *os.FileInfo) {
	file_info, err := os.Stat(file_path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
	}
	return true, &file_info
}

// ReadLines reads contents from file and splits them by new line.
// The offset tells at which line number to start.
// The count determines the number of lines to read (starting from offset):
//   n >= 0: at most n lines
//   n < 0: whole file
func readlines_offset_linecount(filename string, offset uint, n int) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []string{""}, err
	}
	defer f.Close()

	var ret []string

	r := bufio.NewReader(f)
	for i := 0; i < n+int(offset) || n < 0; i++ {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		if i < int(offset) {
			continue
		}
		ret = append(ret, strings.Trim(line, "\n"))
	}
	return ret, nil
}
