/*
 * @Author: calmwu
 * @Date: 2017-11-18 12:11:14
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-11-18 16:25:01
 */

package kvdata

import (
	crand "crypto/rand"
	"fmt"
	"sync"
	"testing"
	"time"

	"sailcraft/base"

	"github.com/hashicorp/consul/api"
)

func NewConsulClient(consulIP string) (*api.Client, error) {
	conf := api.DefaultConfig()
	if consulIP != "127.0.0.1" {
		conf.Address = fmt.Sprintf("%s:8500", consulIP)
	}
	return api.NewClient(conf)
}

func testKey() string {
	buf := make([]byte, 16)
	if _, err := crand.Read(buf); err != nil {
		panic(fmt.Errorf("Failed to read random bytes: %v", err))
	}

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
		buf[0:4],
		buf[4:6],
		buf[6:8],
		buf[8:10],
		buf[10:16])
}

func ConsulKVPut(client *api.Client, kvPair *api.KVPair) error {
	kv := client.KV()
	_, err := kv.Put(kvPair, nil)
	return err
}

func ConsulKVGet(client *api.Client, keyName string, qOpt *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error) {
	kv := client.KV()
	kvPair, qMeta, err := kv.Get(keyName, qOpt)
	fmt.Printf("kvPair:%+v qMeta:%+v\n", kvPair, qMeta)
	return kvPair, qMeta, err
}

func TestKVGet(t *testing.T) {
	client, err := NewConsulClient("10.135.138.179")
	if err != nil {
		t.Error(err.Error())
		return
	}

	kv := client.KV()
	kvPair, qMeta, err := kv.Get("ryzen", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}

	t.Logf("kvPair:%+v qMeta:%+v value:[%s]", kvPair, qMeta, string(kvPair.Value))
}

func TestKVSetGet(t *testing.T) {
	client, err := NewConsulClient("10.135.138.179")
	if err != nil {
		t.Error(err.Error())
		return
	}

	value := fmt.Sprintf("%s", testKey())
	p := &api.KVPair{Key: "ryzen", Flags: 42, Value: []byte(value)}

	kv := client.KV()
	wMeta, err := kv.Put(p, nil)
	if err != nil {
		t.Error(err.Error())
		return
	}

	t.Logf("wMeta:%v", wMeta)

	kvPair, qMeta, err := kv.Get("ryzen", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}

	t.Logf("kvPair:%+v qMeta:%+v value:[%s]", kvPair, qMeta, string(kvPair.Value))
}

func TestKVParallelGet(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			client, err := NewConsulClient("10.135.138.179")
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			kvPair, qMeta, err := client.KV().Get("ryzen", nil)
			if err != nil {
				t.Error(err.Error())
				return
			}

			fmt.Printf("kvPair:%+v qMeta:%+v value:[%s]\n", kvPair, qMeta, string(kvPair.Value))
		}()
	}

	wg.Wait()
}

func TestKVCas(t *testing.T) {
	client, err := NewConsulClient("10.135.138.179")
	if err != nil {
		t.Error(err.Error())
		return
	}

	value := fmt.Sprintf("%s", testKey())
	kp := &api.KVPair{Key: "ryzen", Flags: 42, Value: []byte(value)}

	kv := client.KV()
	b, wMeta, err := kv.CAS(kp, nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	// 这里肯定修改失败，因为没有设置ModifyIndex
	t.Logf("b[%v] wMeta:%+v", b, wMeta)

	// 重新获取
	kp, qMeta, err := kv.Get("ryzen", nil)
	// 这里有点疑惑，难道这两个值还有不一样的时候
	kp.ModifyIndex = qMeta.LastIndex
	kp.Value = []byte("Fuck sb chehua")
	b, wMeta, err = kv.CAS(kp, nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	// 这里肯定修改失败，因为没有设置ModifyIndex
	t.Logf("b[%v] wMeta:%+v", b, wMeta)
}

func TestGetWithoutKey(t *testing.T) {
	client, err := NewConsulClient("10.135.138.179")
	if err != nil {
		t.Error(err.Error())
		return
	}

	kv := client.KV()
	_, qMeta, err := kv.Get("without", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}

	// 这里会发现，lastIndex是个全局变量，应该是下次设置的值
	t.Logf("qMeta:%+v", qMeta)
}

func TestKVKeyWatch(t *testing.T) {
	client, err := NewConsulClient("118.89.34.64")
	if err != nil {
		t.Error(err.Error())
		return
	}

	kv := client.KV()

	t.Logf("time:%s", time.Now().String())
	pair, meta, err := kv.Get("NotExist", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Logf("time:%s pair:%+v meta:%+v", time.Now().String(), pair, meta)

	lastIndex := meta.LastIndex

	for {
		// 阻塞于此监控key的最新操作，包括删除、创建、更新，Get都会及时返回。
		// 但是子节点的创建与其无关
		qOption := &api.QueryOptions{WaitIndex: lastIndex}
		t.Logf("start time:%s", time.Now().String())
		pair, meta, err = kv.Get("NotExist", qOption)

		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Printf("pair:%+v, meta:%+v\n", pair, meta)
			lastIndex = meta.LastIndex
		}
	}

	t.Logf("end time:%s", time.Now().String())
}

func TestKVAcquire(t *testing.T) {
	client, err := NewConsulClient("10.135.138.179")
	if err != nil {
		t.Error(err.Error())
		return
	}

	acquireKeyName := "kvacquire"

	kv := client.KV()
	session := client.Session()

	sessionEntry := &api.SessionEntry{
		Name: acquireKeyName,
		//Behavior:  api.SessionBehaviorDelete, // 调用release后，session关联的key就会被删除
		// ttl，session的过期时间，实际的过期时间是ttl的两倍，session过期后行为由behavior确定，到底是删除所有绑定的key
		// 还是仅仅解锁
		TTL:       "10s",
		LockDelay: 10 * time.Second,
	}
	sessionID, _, err := session.Create(sessionEntry, nil)

	sessionInfo, qMeta, err := session.Info(sessionID, nil)

	fmt.Printf("------sessionInfo:%+v qMeta:%+v\n", sessionInfo, qMeta)

	// lckSessionID, _, err := session.CreateNoChecks(nil, nil)
	// if err != nil {
	// 	t.Error(err.Error())
	// 	return
	// }

	kp := &api.KVPair{Key: acquireKeyName, Value: []byte(acquireKeyName), Session: sessionID}
	ok, _, err := kv.Acquire(kp, nil)
	if err != nil {
		t.Error(err.Error())
		return
	}

	if ok {
		fmt.Printf("%s Acquire key[%s] successed!\n", base.GetTimeStampMs(), acquireKeyName)
	} else {
		fmt.Printf("Acquire key[%s] not obtain!\n", acquireKeyName)
	}

	kvPair, _, _ := ConsulKVGet(client, acquireKeyName, nil)
	rStr, _ := base.RandomBytes(32)
	kvPair.Value = rStr
	// 自己持有的是可以修改的
	ConsulKVPut(client, kvPair)

	time.Sleep(20 * time.Second)
	fmt.Println("ttl timeout, session behavior release!")
	time.Sleep(20 * time.Second)

	ok, _, err = kv.Release(kp, nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	} else {
		fmt.Printf("%s Release key[%s] successed!", base.GetTimeStampMs(), acquireKeyName)
	}

	ConsulKVGet(client, acquireKeyName, nil)
}
