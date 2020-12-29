package redistool

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/sysconf"
)

const (
	REDIS_SESSION_COUNT = 20
)

type RedisManager struct {
	redisCluster   *RedisCluster
	singletonRedis *RedisNode
}

var (
	GRedisManager = new(RedisManager)
)

func (redisManager *RedisManager) GetClusterRedisMgr(key string) (*RedisNode, error) {
	if key != "" {
		node, err := redisManager.redisCluster.GetRedisNodeByKey(key)
		if err != nil {
			return nil, fmt.Errorf("cluster redisMgr is not exist")
		}

		return node, nil
	}

	return nil, fmt.Errorf("key is empty")
}

func (redisManager *RedisManager) GetSingletonRedisMgr() (*RedisNode, error) {
	if redisManager.singletonRedis != nil {
		return redisManager.singletonRedis, nil
	}

	return nil, fmt.Errorf("SingletonRedis is nil")
}

func newRedisMgr(addr string) *RedisNode {
	redisMgr := NewRedis(addr, REDIS_SESSION_COUNT)
	if redisMgr != nil {
		redisMgr.Start()
	}

	return redisMgr
}

func (redisManager *RedisManager) Initialize() error {
	base.GLog.Debug("RedisManager Init enter")

	clusterRedisAddrList := sysconf.GRedisConfig.ClusterRedisAddressList
	singletonRedisAddr := sysconf.GRedisConfig.SingletonRedisAddrsss

	redisManager.singletonRedis = newRedisMgr(singletonRedisAddr)

	redisNodes := make([]*RedisNode, 0)
	for _, addr := range clusterRedisAddrList {
		clusterRedis := newRedisMgr(addr)
		redisNodes = append(redisNodes, clusterRedis)
	}

	cluster, err := GetRedisCluster(redisNodes)
	if err != nil {
		return err
	}

	redisManager.redisCluster = cluster

	return nil
}
