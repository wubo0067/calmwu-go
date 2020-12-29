/*
 * @Author: calmwu
 * @Date: 2017-11-10 15:05:45
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-12-27 15:38:22
 * @Comment:
 */

package redistool

import (
	"errors"
	"fmt"
	"sailcraft/base"
	"strconv"

	"github.com/fzzy/radix/redis"
)

const (
	REDIS_CLUSTER_SLOT_COUNT = 16384
)

type RedisCluster struct {
	redisSlots   []*RedisClusterSlotS
	redisNodeMap map[string]*RedisNode
}

func key(arg interface{}) (string, error) {
	switch arg := arg.(type) {
	case int:
		return strconv.Itoa(arg), nil
	case int64:
		return strconv.Itoa(int(arg)), nil
	case float64:
		return strconv.FormatFloat(arg, 'g', -1, 64), nil
	case string:
		return arg, nil
	case []byte:
		return string(arg), nil
	default:
		return "", fmt.Errorf("key: unknown type %T", arg)
	}
}

func hash(key string) uint16 {
	var s, e int
	for s = 0; s < len(key); s++ {
		if key[s] == '{' {
			break
		}
	}

	if s == len(key) {
		return base.Crc16(key) & (REDIS_CLUSTER_SLOT_COUNT - 1)
	}

	for e = s + 1; e < len(key); e++ {
		if key[e] == '}' {
			break
		}
	}

	if e == len(key) || e == s+1 {
		return base.Crc16(key) & (REDIS_CLUSTER_SLOT_COUNT - 1)
	}

	return base.Crc16(key[s+1:e]) & (REDIS_CLUSTER_SLOT_COUNT - 1)
}

// 传入key得到槽位
func getSlotByKey(rKey interface{}) (uint16, error) {
	key, err := key(rKey)
	if err != nil {
		base.GLog.Error("key %+v is invalid! reason[%s]", rKey, err.Error())
		return 0, err
	}

	slot := hash(key)
	return slot, nil
}

func GetRedisCluster(redisNodes []*RedisNode) (*RedisCluster, error) {
	rc := new(RedisCluster)

	// 通过命令得到cluster的分布情况 info, err := Values(node.do("CLUSTER", "SLOTS"))
	if len(redisNodes) == 0 {
		return nil, errors.New("RedisNode is empty!")
	}

	// 获取cluster分布情况
	slots, err := redisNodes[0].ClusterSlots()
	if err != nil {
		return nil, err
	}

	rc.redisSlots = slots
	rc.redisNodeMap = make(map[string]*RedisNode)
	for i := range redisNodes {
		rc.redisNodeMap[redisNodes[i].RedisSvrAddr] = redisNodes[i]
	}

	for _, slot := range slots {
		if _, ok := rc.redisNodeMap[slot.RedisSvrAddr]; !ok {
			redisNode := newRedisMgr(slot.RedisSvrAddr)
			rc.redisNodeMap[slot.RedisSvrAddr] = redisNode
		}
	}

	return rc, nil
}

func (rc *RedisCluster) GetRedisNodeByKey(rKey interface{}) (*RedisNode, error) {
	// 通过key得到槽位
	slot, err := getSlotByKey(rKey)
	if err != nil {
		return nil, err
	}

	// 比对slot和cluster的分布
	var redisNodeAddr string
	for _, clusterSlot := range rc.redisSlots {
		if uint16(clusterSlot.BeginPos) <= slot && slot <= uint16(clusterSlot.EndPos) {
			redisNodeAddr = clusterSlot.RedisSvrAddr
		}
	}

	//
	if redisNode, ok := rc.redisNodeMap[redisNodeAddr]; ok {
		base.GLog.Debug("rKey:%v slot:[%d] redisNodeAddr[%s]", rKey, slot, redisNodeAddr)
		return redisNode, nil
	}

	return nil, fmt.Errorf("rKey:%v slot:[%d] redisNodeAddr[%s] is invalid!", rKey, slot, redisNodeAddr)
}

func redisClusterSlotsGet(conn *redis.Client, redisCmdData *RedisCommandData) ([]*RedisClusterSlotS, error) {
	reply := conn.Cmd("CLUSTER", "SLOTS")
	if reply.Err == nil {
		slotCount := len(reply.Elems)
		clusterSlots := make([]*RedisClusterSlotS, slotCount)

		for i, e := range reply.Elems {
			clusterSlot := new(RedisClusterSlotS)
			clusterSlot.BeginPos, _ = strconv.ParseInt(e.Elems[0].String(), 10, 64)
			clusterSlot.EndPos, _ = strconv.ParseInt(e.Elems[1].String(), 10, 64)

			clusterSlot.NodeInfo.IP = e.Elems[2].Elems[0].String()
			clusterSlot.NodeInfo.Port, _ = strconv.ParseInt(e.Elems[2].Elems[1].String(), 10, 64)
			clusterSlot.NodeInfo.Key = e.Elems[2].Elems[2].String()

			clusterSlot.RedisSvrAddr = fmt.Sprintf("%s:%d", clusterSlot.NodeInfo.IP, clusterSlot.NodeInfo.Port)

			clusterSlots[i] = clusterSlot
		}
		return clusterSlots, nil
	}
	return nil, reply.Err
}
