/*
 * @Author: calmwu
 * @Date: 2017-11-23 15:36:57
 * @Last Modified by: calmwu
 * @Last Modified time: 2020-08-15 20:14:04
 * @Comment:
 */

package consulapi

import (
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/wubo0067/calmwu-go/utils"
)

func ConsulGlobalSeq(client *api.Client, seqName string, initSeqNum int, waitTime string) (int, error) {

	start := time.Now()
	lockWaitTime, err := time.ParseDuration(waitTime)
	if err != nil {
		utils.ZLog.Errorf("SeqName[%s] parse waitTime failed! reason[%s]", seqName, err.Error())
		return 0, err
	}
	session := client.Session()
	sessionEntry := &api.SessionEntry{
		Name:     seqName,
		TTL:      api.DefaultLockSessionTTL,  // 系统默认的 session ttl 过期时间是实际 ttl * 2
		Behavior: api.SessionBehaviorRelease, // 系统默认行为
	}

	sessionID, _, err := session.Create(sessionEntry, nil)
	if err != nil {
		utils.ZLog.Errorf("SeqName[%s] session create failed! reason[%s]", seqName, err.Error())
		return 0, err
	}

	qOption := &api.QueryOptions{
		WaitTime: lockWaitTime,
	}

	currSeqNum := initSeqNum
	kv := client.KV()

	for {
		// 判断是否有 key 是否有 session 绑定，如果有就需要等待释放
		elapsed := time.Since(start)

		if elapsed > lockWaitTime {
			// 等待超时
			utils.ZLog.Warnf("seqName[%s] wait timeout!", seqName)
			return 0, fmt.Errorf("seqName[%s] wait timeout", seqName)
		}

		qOption.WaitTime -= elapsed
		pair, meta, err := kv.Get(seqName, qOption)
		if err != nil {
			utils.ZLog.Errorf("SeqName[%s] get failed! reason[%s]", seqName, err.Error())
			return 0, err
		}

		//fmt.Printf("---pair:%+v, meta:%+v\n", pair, meta)

		// if pair != nil && pair.Flags != api.LockFlagValue {
		// 	return 0, api.ErrLockConflict
		// }

		if pair != nil && pair.Session == sessionID {
			// 已经持有 lock
			break
		}

		if pair != nil && pair.Session != "" {
			// 有其它 session 绑定了 seqName，需要等待
			qOption.WaitIndex = meta.LastIndex
			//fmt.Println("++++++++++++")
			continue
		}

		if pair == nil {
			pair = &api.KVPair{
				Key:     seqName,
				Value:   []byte("strconv.Itoa(initSeqNum)"),
				Session: sessionID,
				Flags:   api.LockFlagValue,
			}
		} else {
			pair.Session = sessionID
			currSeqNum, _ = strconv.Atoi(string(pair.Value))
		}

		// 尝试 acquire
		//fmt.Printf("acquire pair:%+v\n", pair)
		locked, _, err := kv.Acquire(pair, nil)
		if err != nil {
			utils.ZLog.Errorf("SeqName[%s] acquire failed! reason[%s]", seqName, err.Error())
			return 0, err
		}

		if !locked {
			//fmt.Printf("SeqName[%s] lock false", seqName)
			// 设置一个最小的 index，这样可以立即返回
			qOption.WaitIndex = 0
			pair, meta, err = kv.Get(seqName, qOption)
			if err == nil && pair != nil && pair.Session != "" {
				qOption.WaitIndex = meta.LastIndex
				continue
			}
		} else {
			break
		}
	}
	// 已经 acquire ok!
	//fmt.Printf("Seq[%s] acquire ok! currSeqNum:%d\n", seqName, currSeqNum)

	nextSeqNum := currSeqNum + 1
	// release 也可以设置值
	defer kv.Release(&api.KVPair{
		Key:     seqName,
		Value:   []byte(strconv.Itoa(nextSeqNum)),
		Session: sessionID,
		Flags:   api.LockFlagValue,
	}, nil)

	return currSeqNum, nil
}
