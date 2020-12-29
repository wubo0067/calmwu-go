/*
 * @Author: calmwu
 * @Date: 2017-10-30 16:58:40
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-11-06 17:09:29
 * @Comment:
 */
package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/hashicorp/memberlist"
)

// ./test --bindaddr=192.168.12.3 --bindport=10080 --join=1 --seed=192.168.12.1
// .\test.exe --bindaddr=192.168.12.1 --bindport=10080
// dlv exec ./test -- --bindaddr=192.168.12.3 --bindport=10080 --join=1 --seed=192.168.12.1

var (
	cmdParamsBindAddr   = flag.String("bindaddr", "192.168.12.1", "")
	cmdParamsBindPort   = flag.Int("bindport", 10080, "")
	cmdParamsJoin       = flag.Int("join", 0, "")
	cmdParamsIsSeedAddr = flag.String("seed", "192.168.12.1", "")
)

func createMemberList(bindAddr string, bindPort int) (*memberlist.Memberlist, error) {
	config := memberlist.DefaultLANConfig()
	config.BindAddr = bindAddr
	config.Name = bindAddr
	config.BindPort = bindPort
	config.EnableCompression = false
	config.ProbeInterval = 2 * time.Second
	config.GossipInterval = time.Second
	newMemberList, err := memberlist.Create(config)
	return newMemberList, err
}

func main() {
	flag.Parse()

	m, err := createMemberList(*cmdParamsBindAddr, *cmdParamsBindPort)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer m.Shutdown()

	if *cmdParamsJoin != 0 {
		num, err := m.Join([]string{*cmdParamsIsSeedAddr})
		if err != nil {
			fmt.Printf(err.Error())
		} else {
			fmt.Printf("num: %d\n", num)

			fmt.Printf("m members %d\n", len(m.Members()))
		}
	}

	if *cmdParamsJoin == 0 {
		for {
			time.Sleep(10 * time.Second)
		}
	} else {
		time.Sleep(60 * time.Second)60 * time.Second)
	}
}
