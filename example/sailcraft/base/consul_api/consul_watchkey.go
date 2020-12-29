/*
 * @Author: calmwu
 * @Date: 2017-11-27 10:25:13
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-11-27 12:32:51
 */

package consul_api

import (
	"errors"
	"fmt"
	"sailcraft/base"
	"time"

	"github.com/hashicorp/consul/api"
)

var (
	ConsulWatchKeyNotExist = errors.New("Watch Key dos not exist")
	ConsulWatchExit        = errors.New("Watch exit")
)

// 对一个key的监控
func ConsulWatchKey(client *api.Client, keyName string, stopCh chan struct{}, notifyCh chan interface{}) error {

	if client == nil {
		return fmt.Errorf("client is nil")
	}

	if len(keyName) == 0 {
		return fmt.Errorf("keyName is empty")
	}

	kv := client.KV()
	pair, meta, err := kv.Get(keyName, nil)
	if err != nil {
		base.GLog.Error("Get key[%s] failed! reason[%s]", keyName, err.Error())
		return err
	}

	prevLastIndex := uint64(0)
	queryOpt := &api.QueryOptions{
		WaitIndex: 0, // 第一次不用等待
		WaitTime:  3 * time.Second,
	}

	go func() {
		for {
			pair, meta, err = kv.Get(keyName, queryOpt)
			if err != nil {
				notifyCh <- err
				base.GLog.Error("Get key[%s] failed! reason[%s]", keyName, err.Error())
			}

			queryOpt.WaitTime = 3 * time.Second
			queryOpt.WaitIndex = meta.LastIndex

			if pair == nil {
				notifyCh <- ConsulWatchKeyNotExist
				base.GLog.Warn("Key[%s] dose not exist!", keyName)
			} else {
				if meta.LastIndex > prevLastIndex {
					// fmt.Printf("push value[%s] into notifyCh, prevLastIndex:%d meta.LastIndex:%d\n",
					// 	string(pair.Value), prevLastIndex, meta.LastIndex)
					notifyCh <- pair.Value
					prevLastIndex = meta.LastIndex
				}
			}

			select {
			case <-stopCh:
				{
					fmt.Println("receive exit notify")
					notifyCh <- ConsulWatchExit
				}
			default:
				break
			}
		}
	}()

	return nil
}
