/*
 * @Author: CALM.WU
 * @Date: 2021-01-27 14:07:32
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2021-01-27 14:32:53
 */

package main

import (
	"fmt"
	"net"
	"os"

	"github.com/florianl/go-tc"
)

func main() {
	// open a rtnetlink socket
	rtnl, err := tc.Open(&tc.Config{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not open rtnetlink socket: %v\n", err)
		return
	}
	defer func() {
		if err := rtnl.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "could not close rtnetlink socket: %v\n", err)
		}
	}()

	// get all the qdiscs from all interfaces
	qdiscs, err := rtnl.Qdisc().Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get qdiscs: %v\n", err)
		return
	}

	for _, qdisc := range qdiscs {
		iface, err := net.InterfaceByIndex(int(qdisc.Ifindex))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Itf: %#v, could not get interface from id %d: %v", iface, qdisc.Ifindex, err)
		} else {
			fmt.Printf("%20s\t%s\n", iface.Name, qdisc.Kind)
		}
	}
}
