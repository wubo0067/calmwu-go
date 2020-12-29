/*
 * @Author: calmwu
 * @Date: 2017-11-21 14:52:04
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-11-21 17:27:18
 * @Comment:
 */
package consul_api

import (
	"fmt"
	"sailcraft/base"
	"time"

	"github.com/hashicorp/consul/api"
)

/*
client:			consul api对象
lockName:		锁名字
watiLockTime:	得到锁的等待时间，如果超时都没有获得，返回报错 "10s" "100ms"，永久等待就用""
*/
func ConsulGlobalLock(client *api.Client, lockName string, watiLockTime string) (*api.Lock, error) {
	if client == nil {
		return nil, fmt.Errorf("CAPI: Consul client is nil")
	}

	if len(lockName) == 0 {
		return nil, fmt.Errorf("CAPI: LockName is empty")
	}

	opts := &api.LockOptions{
		Key: lockName,
	}

	if len(watiLockTime) > 0 {
		lockWaitDuration, err := time.ParseDuration(watiLockTime)
		if err != nil {
			return nil, err
		}
		// 锁的等待时间
		opts.LockTryOnce = true
		opts.LockWaitTime = lockWaitDuration
	}

	lock, err := client.LockOpts(opts)
	if err != nil {
		base.GLog.Error(err.Error())
		return nil, err
	}

	base.GLog.Debug("lockName[%s] watiLockTime[%s]", lockName, watiLockTime)

	// 加锁
	leaderCh, err := lock.Lock(nil)
	if err != nil {
		base.GLog.Error("Lock %s failed! reason[%s]", lockName, err.Error())
		return nil, err
	}

	if leaderCh == nil && err == nil {
		base.GLog.Error("Lock %s acquire timeout", lockName)
		return nil, fmt.Errorf("CAPI: Acquire Lock timeout")
	}

	go func() {
		select {
		case <-leaderCh:
			base.GLog.Warn("Lock %s not leader, must be check!!!")
		}
	}()

	return lock, nil
}

func ReleaseGlobalLock(lock *api.Lock) error {
	if lock == nil {
		return fmt.Errorf("lock is nil")
	}

	err := lock.Unlock()
	if err != nil {
		base.GLog.Error("Lock release failed! reason[%s]", err.Error())
	}

	err = lock.Destroy()
	if err != nil {
		base.GLog.Error("Lock destroy failed! reason[%s]", err.Error())
	}

	return err
}
