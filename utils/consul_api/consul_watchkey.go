/*
 * @Author: calmwu
 * @Date: 2017-11-27 10:25:13
 * @Last Modified by: calmwu
 * @Last Modified time: 2020-08-15 20:18:02
 */

package consulapi

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"github.com/wubo0067/calmwu-go/utils"
)

var (
	ErrConsulWatchKeyNotExist = errors.New("watch Key dos not exist")
	ErrConsulWatchExit        = errors.New("watch exit")
)

// ConsulWatchKey 对一个key的监控
func ConsulWatchKey(client *api.Client, keyName string, stopCh chan struct{}, notifyCh chan interface{}) error {
	if client == nil {
		return fmt.Errorf("client is nil")
	}

	if keyName == "" {
		return fmt.Errorf("keyName is empty")
	}

	kv := client.KV()
	pair, meta, err := kv.Get(keyName, nil)
	if err != nil {
		utils.ZLog.Errorf("Get key[%s] failed! reason[%s]", keyName, err.Error())
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
				utils.ZLog.Errorf("Get key[%s] failed! reason[%s]", keyName, err.Error())
			}

			queryOpt.WaitTime = 3 * time.Second
			queryOpt.WaitIndex = meta.LastIndex

			if pair == nil {
				notifyCh <- ErrConsulWatchKeyNotExist
				utils.ZLog.Warn("Key[%s] dose not exist!", keyName)
			} else if meta.LastIndex > prevLastIndex {
				// fmt.Printf("push value[%s] into notifyCh, prevLastIndex:%d meta.LastIndex:%d\n",
				// 	string(pair.Value), prevLastIndex, meta.LastIndex)
				notifyCh <- pair.Value
				prevLastIndex = meta.LastIndex
			}

			select {
			case <-stopCh:
				{
					fmt.Println("receive exit notify")
					notifyCh <- ErrConsulWatchExit
					return
				}
			default:
				break
			}
		}
	}()

	return nil
}
