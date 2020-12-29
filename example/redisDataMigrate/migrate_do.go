package main

import (
	"strings"
	"sync"
	"time"
)

const (
	E_REDIS_CONN_DIALTIMEOUT = 10
)

func DoMigrate(config *Config) {

	begin_time := time.Now().String()
	// 启动
	gMigrateWorkMgr.StartWorks(config)

	var gatherKeysWaitGroup *sync.WaitGroup = new(sync.WaitGroup)
	for index, _ := range config.MigrateFromServs {
		FromRedis := &config.MigrateFromServs[index]
		// 对每个机器采用一个routine进行key采集
		gatherKeysWaitGroup.Add(1)
		go GatherRedisKeys(FromRedis, config.ScanCount, gatherKeysWaitGroup)
	}
	gatherKeysWaitGroup.Wait()

	gLog.Debug("------------------------------------------------------------------")

	// 发送结束标志位
	gMigrateWorkMgr.StopWorks(config)
	// 等待结束
	gMigrateWorkMgr.WaitWorks()

	end_time := time.Now().String()
	gLog.Debug("DoMigrate completed! begin_time[%s] end_time[%s]", begin_time, end_time)
}

func GatherRedisKeys(FromRedis *RedisSvrInfo, scanCount int, gatherKeysWaitGroup *sync.WaitGroup) {
	defer gatherKeysWaitGroup.Done()

	redisAddr := FromRedis.RedisSvrAddr
	gLog.Debug("start Migrate redis[%s] data", redisAddr)

	// 获取连接对象
	connRedis := gRedisConnPoolMgr.GetRedisConn(redisAddr)
	if connRedis == nil {
		return
	}
	defer gRedisConnPoolMgr.PutRedisConn(redisAddr, connRedis)

	if scanCount <= 0 {
		gLog.Error("Config ScanCount[%d] is invalid!", scanCount)
		return
	}

	// 获得数据库中key的数量
	keyCount, err := connRedis.Cmd("DBSIZE").Int()
	if err != nil {
		gLog.Debug("Redis[%s] DBSIZE failed! error[%s]", redisAddr, err.Error())
		return
	}

	gLog.Debug("Redis[%s] key count[%d]", redisAddr, keyCount)

	// 循环scan key
	var scanCursor int64 = 0
	for keyCount > 0 {
		//scanCmd := fmt.Sprintf("SCAN %d COUNT %d", scanCursor, scanCount)
		reply := connRedis.Cmd("SCAN", scanCursor, "COUNT", scanCount)
		if reply.Err != nil {
			gLog.Error("Redis[%s] SCAN cursor[%d] failed! err[%s]", redisAddr, scanCursor, reply.Err.Error())
			return
		}
		scanCursor, _ = reply.Elems[0].Int64()
		keyList, err := reply.Elems[1].List()
		if err != nil {
			gLog.Error("Redis[%s] scan keyList failed! err[%s]", redisAddr, err.Error())
			return
		}
		//gLog.Debug("%v", keyList)

		// 得到每个key的类型
		for _, key := range keyList {
			// 得到key的类型
			keyType := connRedis.Cmd("TYPE", key).String()
			keyTTL, _ := connRedis.Cmd("TTL", key).Int()
			//gLog.Debug("Redis[%s] key[%s] type[%s]", redisAddr, key, keyType)
			if !strings.EqualFold(keyType, "zset") {
				migrateInfo := new(MigrateInfo)
				migrateInfo.FromRedisAddr = redisAddr
				migrateInfo.IsFinal = false
				migrateInfo.RedisKey = key
				migrateInfo.RedisKeyType = keyType
				migrateInfo.RedisKeyTTL = keyTTL
				gMigrateWorkMgr.DispatchMigrateInfo(migrateInfo)
			} else {
				gLog.Warn("RedisSvr[%s] key[%s] type is zset, no need to migration",
					redisAddr, key)
			}
		}
		keyCount -= len(keyList)
		//time.Sleep(time.Millisecond * 500)
		gLog.Debug("remainder_key count[%d]", keyCount)
	}

	return
}
