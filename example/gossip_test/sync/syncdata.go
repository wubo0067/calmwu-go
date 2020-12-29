/*
 * @Author: calmwu
 * @Date: 2017-11-06 14:43:15
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-11-06 17:09:30
 * @Comment:
 */

package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/hashicorp/memberlist"
)

// .\syncdata --bindaddr=192.168.12.1 --bindport=10080 --seed=192.168.12.3
// ./syncdata --bindaddr=192.168.12.3

var (
	cmdParamsBindAddr = flag.String("bindaddr", "192.168.12.1", "")
	cmdParamsBindPort = flag.Int("bindport", 10080, "")
	cmdParamsSeedAddr = flag.String("seed", "", "")
)

type ClientDelegateS struct {
	meta        []byte // 用户自定义的节点元数据，保存在node节点中
	msgs        [][]byte
	broadcasts  [][]byte
	state       []byte
	remoteState []byte
}

// 这不是更新后的通知
func (m *ClientDelegateS) NodeMeta(limit int) []byte {
	fmt.Printf("memberlist get nodeMeta, limit:%d\n", limit)
	return m.meta
}

func (m *ClientDelegateS) NotifyMsg(msg []byte) {
	fmt.Printf("NotifyMsg be invoked\n")
	cp := make([]byte, len(msg))
	copy(cp, msg)
	// 消息会堆积
	m.msgs = append(m.msgs, cp)
}

func (m *ClientDelegateS) GetBroadcasts(overhead, limit int) [][]byte {
	b := m.broadcasts
	m.broadcasts = nil
	return b
}

func (m *ClientDelegateS) LocalState(join bool) []byte {
	return m.state
}

func (m *ClientDelegateS) MergeRemoteState(s []byte, join bool) {
	m.remoteState = s
}

func createMemberList(bindAddr string,
	bindPort int,
	eventChan chan memberlist.NodeEvent,
	clientDeletegate memberlist.Delegate) (*memberlist.Memberlist, error) {
	config := memberlist.DefaultLANConfig()
	config.BindAddr = bindAddr
	config.Name = bindAddr
	config.BindPort = bindPort
	config.EnableCompression = false
	// 这个默认值是1秒
	//config.ProbeInterval = 2 * time.Second
	//config.GossipInterval = time.Second
	// 不要用tcpping，这样可以节省超时时间
	config.DisableTcpPings = true

	if *cmdParamsSeedAddr == "" && eventChan != nil {
		config.Events = &memberlist.ChannelEventDelegate{eventChan}
	}

	if clientDeletegate != nil {
		config.Delegate = clientDeletegate
	}

	newMemberList, err := memberlist.Create(config)
	return newMemberList, err
}

func main() {
	flag.Parse()

	eventChan := make(chan memberlist.NodeEvent, 16)
	clientDeletegate := new(ClientDelegateS)

	m, err := createMemberList(*cmdParamsBindAddr, *cmdParamsBindPort, eventChan, clientDeletegate)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	defer m.Shutdown()

	if *cmdParamsSeedAddr == "" {
		// 主机
		fmt.Println("this is seed server!")

		for {
			select {
			case e := <-eventChan:
				fmt.Printf("receive event\n\t")
				switch e.Event {
				case memberlist.NodeJoin:
					fmt.Printf("node:%+v join\n", e.Node)
				case memberlist.NodeLeave:
					fmt.Printf("Node:%+v leave\n", e.Node)
				case memberlist.NodeUpdate:
					fmt.Printf("Node name[%s] addr[%s] update\n", e.Node.Name, e.Node.Addr.String())
				}
			}
		}
	} else {
		// 这里要join
		m.Join([]string{*cmdParamsSeedAddr})
		// 创建定时器
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		var counter int
		for {
			select {
			case <-ticker.C:
				{
					// 更新meta
					metaStr := fmt.Sprintf("Meta_%d", counter)
					counter++
					clientDeletegate.meta = []byte(metaStr)
					m.UpdateNode(0)
					fmt.Println("update node meta")
				}
			}
		}
	}

}
