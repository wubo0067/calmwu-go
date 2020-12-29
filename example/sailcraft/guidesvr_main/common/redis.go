/*
 * @Author: calmwu
 * @Date: 2017-12-27 15:19:15
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-12-27 15:26:35
 * @Comment:
 */

package common

import (
	"sailcraft/base"
	"sailcraft/dataaccess/redistool"
	"sailcraft/sysconf"
	"sync"
)

var (
	GRedisCluster *redistool.RedisCluster = nil
	initOnce      sync.Once
)

func InitRedisCluster(redisSvrAddrs []string) error {
	var err error

	initOnce.Do(func() {

		redisNodes := make([]*redistool.RedisNode, len(sysconf.GRedisConfig.ClusterRedisAddressList))
		for index, addr := range sysconf.GRedisConfig.ClusterRedisAddressList {
			redisNodes[index] = redistool.NewRedis(addr, 5)
			err = redisNodes[index].Start()
			if err != nil {
				base.GLog.Error("start redis[%s] failed! reason[%s]", addr, err.Error())
			}
		}

		GRedisCluster, err = redistool.GetRedisCluster(redisNodes)
		if err != nil {
			base.GLog.Error("get redisCluster failed! reason[%s]", err.Error())
		}

		base.GLog.Debug("InitRedisCluster successed!")
	})
	return err
}
