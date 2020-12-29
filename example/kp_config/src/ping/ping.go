package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// lookup host ip address
func Lookup(host string) (string, error) {
	addrs, err := net.LookupHost(host)
	if err != nil {
		return "", err
	}
	if len(addrs) < 1 {
		return "", errors.New("unknown host")
	}
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	return addrs[rd.Intn(len(addrs))], nil
}

// return []byte len(size)
var Data = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/")

type Reply struct {
	Time int64
	TTL  uint8
}

// ping ip4
func Run(addr string, req, timeout int, data []byte) (*Reply, error) {
	// icmp data
	xid, xseq := os.Getpid()&0xffff, req
	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: xid, Seq: xseq,
			Data: data,
		},
	}
	wb, err := wm.Marshal(nil)
	if err != nil {
		return nil, err
	}
	// user must be root/administrators
	c, err := net.Dial("ip4:icmp", addr)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: %s", addr, err))
	}
	defer c.Close()

	c.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
	start := time.Now()

	if _, err := c.Write(wb); err != nil {
		return nil, errors.New(fmt.Sprintf("%s: %s", addr, err))
	}

	rb := make([]byte, 1500)

	n, err := c.Read(rb)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: %s", addr, err))
	}
	duration := time.Now().Sub(start)
	ttl := uint8(rb[8])
	rb = func(b []byte) []byte {
		if len(b) < 20 {
			return b
		}
		hdrlen := int(b[0]&0x0f) << 2
		return b[hdrlen:]
	}(rb)

	rm, err := icmp.ParseMessage(1, rb[:n])
	if err != nil {
		return nil, err
	}

	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		t := int64(duration / time.Millisecond)
		return &Reply{t, ttl}, nil
	case ipv4.ICMPTypeDestinationUnreachable:
		return nil, errors.New(fmt.Sprintf("%s: Destination Unreachable", addr))
	default:
		return nil, errors.New(fmt.Sprintf("not ICMPTypeEchoReply %v", rm))
	}
}

func main() {
	host := "58.63.236.248"
	addr, err := Lookup(host)
	if err != nil {
		fmt.Println(err)
		return
	}
	count := 5
	timeout := 1
	data := Data

	if count < 1 {
		count = 4
	}

	fmt.Printf("ping %s with %d bytes of data:\n", host, len(data))
	for i := 0; i < count; i++ {
		time.Sleep(1 * time.Second)
		r, err := Run(addr, i, timeout, data)
		if err != nil {
			fmt.Println(err)
			continue
		}
		t := fmt.Sprintf("%dms", r.Time)
		// if r.Time < 1 {
		// 	t = fmt.Sprintf("<1ms")
		// }
		fmt.Printf("reply %d from %s: bytes=%d time=%s ttl=%d\n",
			i+1, addr, len(data), t, r.TTL)
	}
}
