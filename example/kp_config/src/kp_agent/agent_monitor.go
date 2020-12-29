package main

import (
	"time"
)

func init_agent_monitor() {
	wait_group.Add(1)
	defer wait_group.Done()

	g_log.Debug("goroutine agent_monitor running!")
	// 检查定时器
	agentcheck_ticker := time.NewTicker(time.Second * 10)
	defer agentcheck_ticker.Stop()
	// 统计定时器
	gather_netinfo_ticker := time.NewTicker(time.Minute * 1)
	defer gather_netinfo_ticker.Stop()

L:
	for {
		select {
		case <-exit_chan:
			g_log.Debug("goroutine agent_monitor receive exit notify!")
			break L
		case <-agentcheck_ticker.C:
			// 检查kpserver的状态
			for key, kpserver_info := range kpserver_map {
				g_log.Debug("kpserver[%s:%s] status[%s]", kpserver_info.m_server_ip,
					kpserver_info.m_server_port, kpserver_info.m_connstate.String())

				if kpserver_info.m_connstate == E_STATE_DISCONNECTED ||
					kpserver_info.m_connstate == E_STATE_INIT {
					start_agent(kpserver_map[key])
				}
			}
		case <-gather_netinfo_ticker.C:
			go do_gather_netinfo(g_kpconfinfo.VrrpInterface)
		}
	}

	g_log.Debug("goroutine agent_monitor exit!")
}
