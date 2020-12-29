package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

// 导入目录后，可以用过package.xx来引用对象
// tmd, ticker.C没法用ticker.stop去关闭 http://stackoverflow.com/questions/17797754/ticker-stop-behaviour-in-golang

var (
	cmd_params_help          = flag.Bool("help", false, "show usage")
	cmd_params_version       = flag.Bool("version", false, "show version and buildtime")
	cmd_params_mode          = flag.String("mode", "", "kp_agent work mode[agent/notify]")
	cmd_params_kpserverinfos = flag.String("kpservs", "", "x.x.x.x:9991,x.x.x.x:9991")
	cmd_params_logpath       = flag.String("logpath", "../log", "logpath")
	cmd_params_local_itf     = flag.String("local_itf", "eth0", "Local Network Interface")
	cmd_params_notify_port   = flag.Int("notify_port", 10001, "rs state change notify udp port")
	cmd_params_rs_ip         = flag.String("rs_ip", "", "real server ip")
	cmd_params_rs_port       = flag.Int("rs_port", 0, "real server port")
	cmd_params_proto_name    = flag.String("proto_name", "", "real server protocol name[tcp/udp]")
	cmd_params_notify_status = flag.String("status", "", "health check status[up/down]")
	kpagent_version          = "0.0.1"
	kpagent_buildtime        = ""
	run_flag                 = true
	// 关闭的通知chan
	exit_chan = make(chan struct{})
	// 同步对象
	wait_group = new(sync.WaitGroup)
	// 本机ip
	host_localIP = ""
)

func show_usage() {
	fmt.Printf("%s\n", "Usage of kp_agent")
	fmt.Printf("\tkp_agent --version\n")
	fmt.Printf("\tkp_agent --help\n")
	fmt.Printf("\tnohup kp_agent --mode=agent --kpservs=ip1:9991,ip2:9991 --local_itf=eth0 --notify_port=10001 &\n")
	fmt.Printf("\tkp_agent --mode=notify --rs_ip=x.x.x.x --rs_port=xxx --proto_name=tcp/udp --status=up/down --notify_port=10001\n")
	return
}

func parse_params() {
	flag.Parse()

	if *cmd_params_help {
		show_usage()
		os.Exit(0)
	} else if *cmd_params_version {
		fmt.Printf("kp_agent version[%s] buildtime[%s]\n", kpagent_version, kpagent_buildtime)
		os.Exit(0)
	}
}

func on_signal() {
	sig_chan := make(chan os.Signal)
	signal.Notify(sig_chan, syscall.SIGINT, syscall.SIGTERM)
L:
	for {
		sig := <-sig_chan
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			g_log.Debug("receive exit signal!")
			//run_flag = false
			close(exit_chan)
			break L
		}
	}
	g_log.Debug("goroutine on_signal exit!")
}

func main() {
	parse_params()

	if strings.Compare(*cmd_params_mode, "agent") == 0 {
		g_log.Debug("mode[agent] kpserver_infos[%s] logpath[%s] local_itf[%s] notfiy_port[%d]\n", *cmd_params_kpserverinfos,
			*cmd_params_logpath, *cmd_params_local_itf, *cmd_params_notify_port)

		if init_log("kp_agent.log") < 0 {
			os.Exit(-1)
		}
		defer g_log.Close()

		g_log.Debug("kp_agent running!")

		g_log.Debug("NumCPU[%d]", runtime.NumCPU())
		runtime.GOMAXPROCS(runtime.NumCPU())

		// 定时器
		//server_check_ticker := time.NewTicker(time.Second)
		//defer server_check_ticker.Stop()

		// 初始化信号
		go on_signal()
		// 初始化状态接受服务
		listener, err := net.ListenUDP("udp", &net.UDPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: *cmd_params_notify_port})
		if err != nil {
			g_log.Debug(err.Error())
			time.Sleep(time.Second)
			os.Exit(-1)
		}
		defer listener.Close()
		go process_rsstatus_notify(listener)

		//
		host_localIP = *get_localip_by_ifname(*cmd_params_local_itf)
		ret := init_agent_net(*cmd_params_kpserverinfos)
		if ret == 0 {
			// 启动监控，通过agent来启动连接
			go init_agent_monitor()
			runtime.Gosched()
			wait_group.Wait()
		}

		g_log.Debug("kp_agent exit!")
		time.Sleep(time.Second)
	} else if strings.Compare(*cmd_params_mode, "notify") == 0 {
		if init_log("kp_notify.log") < 0 {
			os.Exit(-1)
		}
		defer g_log.Close()

		g_log.Debug("mode[notify] rs_ip[%s] rs_port[%d] proto_name[%s] status[%s] notfiy_port[%d]\n", *cmd_params_rs_ip,
			*cmd_params_rs_port, *cmd_params_proto_name, *cmd_params_notify_status, *cmd_params_notify_port)

		notify_rs_status()

		time.Sleep(time.Second)
	} else {
		show_usage()
	}
}
