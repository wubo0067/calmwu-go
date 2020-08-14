/*
 * @Author: calmwu
 * @Date: 2019-02-21 10:20:24
 * @Last Modified by: calmwu
 * @Last Modified time: 2019-02-21 14:41:55
 */

package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/wubo0067/calmwu-go/utils"
	consulapi "github.com/wubo0067/calmwu-go/utils/consul_api"
)

var wg sync.WaitGroup
var logger *log.Logger

func competitiveLock(c *api.Client, lockHeldSecs uint32, id int) {
	defer wg.Done()

	opts := &api.LockOptions{
		Key:         "consol_api/competitiveLock",
		Value:       []byte(fmt.Sprintf("competitiveLock_%d", id)),
		SessionName: fmt.Sprintf("consol_api/competitiveLock_%d", id),
		// LockTryOnce: true,
	}

	lock, err := c.LockOpts(opts)
	if err != nil {
		logger.Printf("LockOpts failed! error:%s\n", err.Error())
		return
	}

	logger.Printf("%d -------before Lock-------\n", id)
	leaderCh, err := lock.Lock(nil)

	if err != nil {
		logger.Printf("%d Lock failed! error:%s\n", id, err.Error())
		return
	}

	// lockSession其实是sessionid
	// 如果leaderCh返回为nil就说明lock抢占没有成功
	logger.Printf("%d lock Info:%#v leaderCh：%v\n", id, lock, leaderCh)

	select {
	case <-time.After(time.Duration(lockHeldSecs) * time.Second):
		break
	case <-leaderCh:
		// 只要手动删除了key，这里就会收到通知
		// 如果停掉consul服务，这里也会收到通知
		logger.Printf("%d receive leaderCh notify\n", id)
		break
	}

	logger.Printf("%d ++++++Unlock-----------\n", id)
	lock.Unlock()
	lock = nil
}

func main() {
	logger = utils.NewSimpleLog(nil)
	consolClient, err := consulapi.NewConsulClient("192.168.2.104")
	if err != nil {
		logger.Printf("New consul client failed! error:%s\n", err.Error())
	}

	logger.Printf("consul Client %v\n", consolClient)

	wg.Add(1)
	go competitiveLock(consolClient, 20, 1)

	wg.Add(1)
	go competitiveLock(consolClient, 20, 2)

	wg.Wait()

	logger.Printf("test exit!\n")
}
