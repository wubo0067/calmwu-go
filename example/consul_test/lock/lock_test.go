/*
 * @Author: calmwu
 * @Date: 2017-11-20 15:29:40
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-11-21 16:44:22
 * @Comment:
 */

package lock

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
)

var r *rand.Rand // Rand for this package.

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func RandomString(strlen int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := range result {
		result[i] = chars[r.Intn(len(chars))]
	}
	return string(result)
}

func NewConsulClient(consulIP string) (*api.Client, error) {
	conf := api.DefaultConfig()
	if consulIP != "127.0.0.1" {
		conf.Address = fmt.Sprintf("%s:8500", consulIP)
	}
	return api.NewClient(conf)
}

func createSession(session *api.Session, lockName string) (string, error) {
	entry := &api.SessionEntry{
		TTL:       "10", // disable ttl
		Name:      lockName,
		LockDelay: 10 * time.Second,
	}
	id, _, err := session.Create(entry, nil)
	return id, err
}

func getLockData(client *api.Client, lockName string) *api.KVPair {
	kv := client.KV()
	pair, meta, err := kv.Get(lockName, nil)
	if err != nil {
		fmt.Printf("err:%v\n", err.Error())
		return nil
	}

	if pair == nil {
		fmt.Printf("key[%s] is not exists!\n", lockName)
	} else {
		fmt.Printf("pair:%+v, meta:%+v, value:[%s]\n", pair, meta, string(pair.Value))
	}
	return pair
}

func putLockData(client *api.Client, kvPair *api.KVPair) {
	kvPair.Value = []byte(RandomString(16))
	kv := client.KV()
	_, err := kv.Put(kvPair, nil)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("----putLockData----")
}

func TestLockRelease(t *testing.T) {
	client, err := NewConsulClient("118.89.34.64")
	if err != nil {
		t.Error(err.Error())
		return
	}

	lockName := "Lock_customdata"

	//session := client.Session()
	//sessionId, err := createSession(session, lockName)

	// 创建锁
	opts := &api.LockOptions{
		Key:   lockName,
		Value: []byte("Hello Lock assoc value!"),
		//Session: sessionId,
		//LockWaitTime: 3 * time.Second,
	}
	lock, err := client.LockOpts(opts)
	if err != nil {
		t.Fatalf("err:%v", err.Error())
	}

	// 加锁
	leaderCh, err := lock.Lock(nil)
	if err != nil {
		t.Fatalf("err:%v", err.Error())
	}

	// 看能否写值，这个flag很重要
	// lockValue := &api.KVPair{Key: lockName, Value: []byte("998"), Flags: api.LockFlagValue}
	// kv := client.KV()
	// _, err = kv.Put(lockValue, nil)
	// if err != nil {
	// 	t.Error(err.Error())
	// }

	exitCh := make(chan struct{})
	go func(client *api.Client, lockName string, leaderCh <-chan struct{}, exitCh chan<- struct{}) {
		// 获取key
		tickerGet := time.NewTicker(1 * time.Second)
		tickerSet := time.NewTicker(3 * time.Second)
		after := time.After(12 * time.Second)
		var pair *api.KVPair
		for {
			select {
			case <-tickerGet.C:
				pair = getLockData(client, lockName)
			case <-tickerSet.C:
				putLockData(client, pair)
			case <-leaderCh:
				// 如果我在管理台上删除了lockName的key，这里会立即收到消息，其实后面有个routine不停的get循环
				fmt.Printf("Recv leaderCh!\n")
				// 解锁之后在看看lockindex session 等信息
				// 解锁之后，session = "", LockIndex没有变化，ModifyIndex递增
				getLockData(client, lockName)
				close(exitCh)
				return
			case <-after:
				//lock, err := client.LockKey(lockName)
				lckOpt := &api.LockOptions{
					Key: lockName,
					// 只有加入下面两项，才能有tryLockTimeOut的情况，lock重试超时后就会返回
					LockWaitTime: 3 * time.Second,
					LockTryOnce:  true,
				}
				lock1, err := client.LockOpts(lckOpt)
				if err != nil {
					fmt.Printf("err:%v\n", err.Error())
				} else {
					fmt.Printf("start:%s\n", time.Now().String())
					// 一直会等在这里，而不是报错
					lCh, err := lock1.Lock(nil)
					if err != nil {
						fmt.Printf("err:%v", err.Error())
					}
					if lCh == nil {
						fmt.Printf("lCh is nil\n")
					}
					fmt.Printf("end:%s\n", time.Now().String())
				}
			}
		}
	}(client, lockName, leaderCh, exitCh)

	// go func() {

	// 	// 真是奇葩了，就算是lock了，其它的请求还是可以修改数据
	// 	// 因为lockindex没了
	// 	for i := 0; i < 3; i++ {
	// 		time.Sleep(2 * time.Second)

	// 		lockValue := &api.KVPair{Key: lockName, Value: []byte("6987"), Flags: api.LockFlagValue}
	// 		kv := client.KV()
	// 		_, err = kv.Put(lockValue, nil)
	// 		if err != nil {
	// 			t.Error(err.Error())
	// 		} else {
	// 			fmt.Println("------------------")
	// 		}
	// 	}

	// }()

	time.Sleep(time.Second * 20)

	// 解锁，解锁之后，leaderCh会被关闭，会打印Recv leaderCh
	err = lock.Unlock()
	if err != nil {
		t.Fatalf("err:%v", err.Error())
	}

	<-exitCh
}
