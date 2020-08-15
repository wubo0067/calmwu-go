/*
 * @Author: calmwu
 * @Date: 2017-11-21 14:52:30
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-29 14:51:21
 * @Comment:
 */

package consulapi

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
)

func TestGlobalLock(t *testing.T) {
	client, err := NewConsulClient("10.135.138.179")
	if err != nil {
		t.Error(err.Error())
		return
	}

	lockName := "GlobalLock/test"

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			var err error
			var lock *api.Lock
			if i > 7 {
				lock, err = ConsulGlobalLock(client, lockName, "4s")
			} else {
				lock, err = ConsulGlobalLock(client, lockName, "")
			}

			if err != nil {
				fmt.Printf("i:%d, %s\n", i, err.Error())
				return
			}

			fmt.Printf("i:%d get lock, %d\n", i, time.Now().Second())

			defer ReleaseGlobalLock(lock)
			defer func() {
				fmt.Printf("i:%d release lock, %d\n", i, time.Now().Second())
			}()
			time.Sleep(time.Second * time.Duration(i))
		}(i)
	}
	wg.Wait()
}

func TestGlobalSeqNum(t *testing.T) {
	t.Parallel()

	seqName := "test_globalseqnum"

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			client, err := NewConsulClient("10.135.138.179")
			if err != nil {
				t.Error(err.Error())
				return
			}

			seqNum, err := ConsulGlobalSeq(client, seqName, 998, "10s")

			if err != nil {
				fmt.Printf("i:%d, %s\n", i, err.Error())
				return
			}

			fmt.Printf("i:%d get seqNum, %d\n", i, seqNum)
		}(i)
	}
	wg.Wait()
}

// dig @10.10.81.214 -p 8600 SailCraft-GuideSvr.service.consul
func TestDns(t *testing.T) {
	client, _ := NewConsulClient("10.10.81.214")

	for i := 0; i < 100; i++ {
		healthServInsts, _ := ConsulServDNS(client, "SailCraft-GuideSvr")
		for index := range healthServInsts {
			fmt.Printf("healthServInsts:%+v\n", healthServInsts[index])
		}
	}
}

func TestWatchKey(t *testing.T) {
	client, _ := NewConsulClient("192.168.68.229")

	stopCh := make(chan struct{})
	notifyCh := make(chan interface{}, 16)

	ConsulWatchKey(client, "recdata-servconf/config", stopCh, notifyCh)

	count := 5
L:
	for {
		select {
		case result := <-notifyCh:
			//fmt.Printf("--------result:%v\n", result)
			if realErr, ok := result.(error); ok {
				if realErr == ErrConsulWatchKeyNotExist {
					fmt.Printf("key[%s] does not exist\n", "WatchKey")
					count--
					if count == 0 {
						close(stopCh)
					}
				} else if realErr == ErrConsulWatchExit {
					fmt.Printf("Watch exit!\n")
					break L
				} else {
					fmt.Printf("error[%s]\n", realErr.Error())
				}
			}

			if data, ok := result.([]byte); ok {
				fmt.Printf("Key[WatchKey] value[%s]\n", string(data))
			}
		}
	}
	fmt.Println("-----------TestWatchKey-----------")
}
