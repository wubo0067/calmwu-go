/*
 * @Author: calmwu
 * @Date: 2017-09-18 14:29:27
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-09-23 11:30:20
 * @Comment:
 */

package data

import (
	"sailcraft/base"
	"sailcraft/indexsvr_main/common"
	"sailcraft/indexsvr_main/proto"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/emirpasic/gods/lists/singlylinkedlist"
	"github.com/fzzy/radix/extra/pool"
	"github.com/fzzy/radix/redis"
)

type RedisDataMgr struct {
	redisConnPools   []*pool.Pool // redis连接池
	bucketCount      int
	RedisDataBuckets []*RedisDataBucket //
}

type RedisDataBucket struct {
	bucketID           int
	dataActionDataChan chan *DataActionInfoS             // 命令队列
	dataMap            map[string]*singlylinkedlist.List // 数据存储
}

func (redisMgr *RedisDataMgr) Load() error {
	if redisMgr.redisConnPools == nil && redisMgr.RedisDataBuckets == nil {

		err := redisMgr.makeConnPools()
		if err != nil {
			return err
		}

		redisMgr.bucketCount = common.GConfig.BucketCount
		redisMgr.RedisDataBuckets = make([]*RedisDataBucket, redisMgr.bucketCount)

		for i := 0; i < redisMgr.bucketCount; i++ {
			redisDataBucket := new(RedisDataBucket)
			redisDataBucket.bucketID = i
			redisDataBucket.dataMap = make(map[string]*singlylinkedlist.List)
			redisDataBucket.dataActionDataChan = make(chan *DataActionInfoS, 1024)

			redisMgr.RedisDataBuckets[i] = redisDataBucket

			go redisDataBucket.redisBucketRun()
		}

		for _, redisConnPool := range redisMgr.redisConnPools {
			err = redisMgr.loadRedisData(redisConnPool)
			if err != nil {
				return err
			}
		}

		// // 做一次查询测试
		// resultList := redisMgr.Like("Captain66", proto.E_DATATYPE_USERINFO)
		// resultList.Each(func(index int, value interface{}) {
		// 	base.GLog.Debug("index[%d] %v", index, value.(proto.DataMetaI))
		// })

		// resultList = redisMgr.Like("no", proto.E_DATATYPE_GUILDINFO)
		// resultList.Each(func(index int, value interface{}) {
		// 	base.GLog.Debug("index[%d] %v", index, value.(proto.DataMetaI))
		// })
	}
	return nil
}

func (redisMgr *RedisDataMgr) makeConnPools() error {

	redisMgr.redisConnPools = make([]*pool.Pool, len(common.GConfig.RedisAddressLst))

	for index, redisAddr := range common.GConfig.RedisAddressLst {
		base.GLog.Debug("Redis host[%s]", redisAddr)

		df := func(network, addr string) (*redis.Client, error) {
			client, err := redis.DialTimeout(network, addr, time.Duration(time.Second))
			if err != nil {
				base.GLog.Error("Connect to Redis[%s] failed! error[%s]", addr, err.Error())
				return nil, err
			}
			return client, nil
		}
		connPool, err := pool.NewCustomPool("tcp", redisAddr, 2, df)
		if err != nil {
			base.GLog.Error("Connecton Redis[%s] failed! %s", redisAddr, base.StrError(err))
			return err
		}

		redisMgr.redisConnPools[index] = connPool
	}

	return nil
}

func (redisMgr *RedisDataMgr) loadRedisData(redisConnPool *pool.Pool) error {
	// 开始读取数据，存入bucket中
	client, err := redisConnPool.Get()
	addr := client.Conn.RemoteAddr()
	if err != nil {
		base.GLog.Error("Get Redis[%s] connection from pool failed! %s", addr, base.StrError(err))
		return err
	} else {
		base.GLog.Debug("Now Load redis data from[%s]", addr)
		defer redisConnPool.Put(client)

		keyCount, err := client.Cmd("DBSIZE").Int()
		if err != nil {
			base.GLog.Debug("Redis[%s] DBSIZE failed! error[%s]", addr, base.StrError(err))
			return err
		}

		base.GLog.Debug("Redis[%s] key count[%d]", addr, keyCount)

		// 循环scan key
		var scanCursor int64 = 0
		for keyCount > 0 {
			reply := client.Cmd("SCAN", scanCursor, "COUNT", 50)
			if reply.Err != nil {
				base.GLog.Error("Redis[%s] SCAN cursor[%d] failed! err[%s]", addr, scanCursor, reply.Err.Error())
				return reply.Err
			}
			scanCursor, _ = reply.Elems[0].Int64()
			keyList, err := reply.Elems[1].List()
			if err != nil {
				base.GLog.Error("Redis[%s] scan keyList failed! err[%s]", addr, err.Error())
				return err
			}
			//gLog.Debug("%v", keyList)

			// 得到每个key的类型
			for _, key := range keyList {
				// 得到key的类型
				keyType := client.Cmd("TYPE", key).String()
				if keyType == "hash" {
					// 如果是hash，判断key是否符合规范
					if strings.Contains(key, common.GConfig.GuildInfo.KeyMatch) {
						// 这是工会的key
						base.GLog.Debug("GuildInfo key[%s]", key)
						record, err := client.Cmd("HGETALL", key).Hash()
						if err != nil {
							base.GLog.Error("HGETALL %s failed! reason[%s]", key, err.Error())
						} else {
							var guildInfo proto.GuildInfoS
							err = mapstructure.Decode(record, &guildInfo)
							if err != nil {
								base.GLog.Error("mapstructure guildInfo key[%s] failed! reason[%s]", key, err.Error())
							} else {
								//base.GLog.Error("record:%v guildInfo:%v", record, guildInfo)
								redisMgr.Set(guildInfo.GuildName, &guildInfo)
								redisMgr.Set(guildInfo.PerformId, &guildInfo)
							}
						}
					} else if strings.Contains(key, common.GConfig.UserInfo.KeyMatch) {
						//
						base.GLog.Debug("UserInfo key[%s]", key)
						record, err := client.Cmd("HGETALL", key).Hash()
						if err != nil {
							base.GLog.Error("HGETALL %s failed! reason[%s]", key, err.Error())
						} else {
							var userInfo proto.UserInfoS
							err = mapstructure.Decode(record, &userInfo)
							if err != nil {
								base.GLog.Error("mapstructure userInfo key[%s] failed! reason[%s]", key, err.Error())
							} else {
								//base.GLog.Error("record:%v userInfo:%v", record, userInfo)
								redisMgr.Set(userInfo.UserName, &userInfo)
								redisMgr.Set(userInfo.Uin, &userInfo)
							}
						}
					}
				}
			}
			keyCount -= len(keyList)
			//time.Sleep(time.Millisecond * 500)
			base.GLog.Debug("remainder_key count[%d]", keyCount)
		}
	}
	return nil
}

// 基于key的模糊查询
func (redisMgr *RedisDataMgr) Like(key string, dataSetType proto.DataSetType, queryCount int) *singlylinkedlist.List {
	return redisMgr.queryData(key, dataSetType, queryCount, E_DATAACTION_LIKE)
}

func (redisMgr *RedisDataMgr) Match(key string, dataSetType proto.DataSetType, queryCount int) *singlylinkedlist.List {
	return redisMgr.queryData(key, dataSetType, queryCount, E_DATAACTION_MATCH)
}

func (redisMgr *RedisDataMgr) queryData(key string, dataSetType proto.DataSetType, queryCount int, actionType DataActionType) *singlylinkedlist.List {
	// 汇总通道
	gatherReplyChan := make(chan *QueryResultS, redisMgr.bucketCount)

	for index := 0; index < redisMgr.bucketCount; index++ {
		redisActionData := new(DataActionInfoS)
		redisActionData.actionType = actionType
		redisActionData.key = key
		redisActionData.resultChan = gatherReplyChan
		redisActionData.dataSetType = dataSetType

		// 并行查询
		redisMgr.RedisDataBuckets[index].dataActionDataChan <- redisActionData
	}

	// 汇总
	var gatherCount = 0
	likeResult := singlylinkedlist.New()
L:
	for {
		select {
		case result, ok := <-gatherReplyChan:
			if ok {
				if result.Ok {
					// 数据汇总
					if !result.Result.Empty() {
						result.Result.Any(
							func(index int, value interface{}) bool {
								if queryCount > 0 {
									likeResult.Append(value.(proto.DataMetaI))
									queryCount--
									return false
								}
								return true
							})
					}
				}
				gatherCount++
				if gatherCount >= redisMgr.bucketCount {
					base.GLog.Debug("queryData key[%s] complete!", key)
					break L
				}
			}
		case <-time.After(3 * time.Second):
			base.GLog.Error("queryData key[%s] timeout!", key)
			// 超时跳出
			break L
		}
	}
	// 返回查询数据
	return likeResult
}

func (redisMgr *RedisDataMgr) Set(key string, data proto.DataMetaI) {
	redisActionData := &DataActionInfoS{
		actionType: E_DATAACTION_SET,
		key:        key,
		value:      data,
	}
	// 计算key的hash值
	hashVal := base.HashStr2Uint32(key)
	pos := hashVal % uint32(redisMgr.bucketCount)
	//base.GLog.Debug("key[%s] hashVal[%d] pos[%d]", key, hashVal, pos)
	redisMgr.RedisDataBuckets[pos].dataActionDataChan <- redisActionData
}

func (redisMgr *RedisDataMgr) Delete(key string, data proto.DataMetaI) {
	redisActionData := &DataActionInfoS{
		actionType: E_DATAACTION_DEL,
		key:        key,
		value:      data,
	}
	// 计算key的hash值
	hashVal := base.HashStr2Uint32(key)
	pos := hashVal % uint32(redisMgr.bucketCount)
	//base.GLog.Debug("key[%s] hashVal[%d] pos[%d]", key, hashVal, pos)

	redisMgr.RedisDataBuckets[pos].dataActionDataChan <- redisActionData
}

func (redisMgr *RedisDataMgr) Modify(oldKey, newKey string, oldData, newData proto.DataMetaI) {
	// 先删除，再SET
	redisMgr.Delete(oldKey, oldData)
	redisMgr.Set(newKey, newData)
}

func (redisMgr *RedisDataMgr) Truncate() {
	redisActionData := &DataActionInfoS{
		actionType: E_DATAACTION_TRUNCATE,
	}
	// 计算key的hash值
	for i := 0; i < redisMgr.bucketCount; i++ {
		redisMgr.RedisDataBuckets[i].dataActionDataChan <- redisActionData
	}
}

func (redisMgr *RedisDataMgr) Reload() error {
	redisMgr.Truncate()
	for _, redisConnPool := range redisMgr.redisConnPools {
		err := redisMgr.loadRedisData(redisConnPool)
		if err != nil {
			return err
		}
	}
	return nil
}

func (redisBucket *RedisDataBucket) redisBucketRun() {
	// 处理命令
	base.GLog.Debug("[%d] redisBucketRun", redisBucket.bucketID)

	for {
		select {
		case action, ok := <-redisBucket.dataActionDataChan:
			if ok {
				switch action.actionType {
				case E_DATAACTION_LIKE:
					base.GLog.Debug("bucketid[%d] doLike key[%s]", redisBucket.bucketID, action.key)
					// 设置返回值
					var queryResult QueryResultS
					queryResult.Ok = true
					queryResult.Result = redisBucket.doLike(action)
					action.resultChan <- &queryResult
					if !queryResult.Result.Empty() {
						base.GLog.Debug("bucketid[%d] key[%s] doLike return!", redisBucket.bucketID, action.key)
					}
				case E_DATAACTION_MATCH:
					base.GLog.Debug("bucketid[%d] doMatch key[%s]", redisBucket.bucketID, action.key)
					var queryResult QueryResultS
					queryResult.Ok = true
					queryResult.Result = redisBucket.doMatch(action)
					action.resultChan <- &queryResult
					if !queryResult.Result.Empty() {
						base.GLog.Debug("bucketid[%d] key[%s] doMatch return!", redisBucket.bucketID, action.key)
					}
				case E_DATAACTION_SET:
					base.GLog.Debug("bucketid[%d] doSet key[%s]", redisBucket.bucketID, action.key)
					redisBucket.doSet(action)
				case E_DATAACTION_TRUNCATE:
					// 创建新的map对象
					base.GLog.Debug("bucketid[%d] doMatch key[%s]", redisBucket.bucketID, action.key)
					redisBucket.dataMap = make(map[string]*singlylinkedlist.List)
				case E_DATAACTION_DEL:
					base.GLog.Debug("bucketid[%d] doDelete key[%s] data[%v]", redisBucket.bucketID, action.key, action.value)
					redisBucket.doDelete(action)
				default:
					base.GLog.Error("bucketid[%d] key[%s] Action cmd[%d] is not support!",
						redisBucket.bucketID, action.key, action.actionType)
				}
			}
		}
	}
}

func (redisBucket *RedisDataBucket) doLike(action *DataActionInfoS) *singlylinkedlist.List {
	// 模糊查找
	resultList := singlylinkedlist.New()
	lowKey := strings.ToLower(action.key)

	for key, dataList := range redisBucket.dataMap {
		// 判断名字是否匹配，部分匹配就可以
		if strings.Contains(key, lowKey) {
			dataList.Each(func(index int, value interface{}) {
				if dataMetaI, ok := value.(proto.DataMetaI); ok {
					if dataMetaI.Type() == action.dataSetType {
						// 这就是要找的数据
						resultList.Append(dataMetaI)
						base.GLog.Debug("bucketid[%d] key[%s] actionKey[%s] index[%d] data:%+v",
							redisBucket.bucketID, key, lowKey, index, dataMetaI)
					}
				}
			})
		}
	}
	return resultList
}

func (redisBucket *RedisDataBucket) doMatch(action *DataActionInfoS) *singlylinkedlist.List {
	resultList := singlylinkedlist.New()
	lowKey := strings.ToLower(action.key)

	for key, dataList := range redisBucket.dataMap {
		// key完全匹配
		if strings.Compare(key, lowKey) == 0 {
			dataList.Each(func(index int, value interface{}) {
				if dataMetaI, ok := value.(proto.DataMetaI); ok {
					if dataMetaI.Type() == action.dataSetType {
						resultList.Append(dataMetaI)
						base.GLog.Debug("bucketid[%d] key[%s] actionKey[%s] index[%d] data:%+v",
							redisBucket.bucketID, key, lowKey, index, dataMetaI)
					}
				}
			})
		}
	}
	return resultList
}

func (redisBucket *RedisDataBucket) doSet(action *DataActionInfoS) {
	// name is key
	lowKey := strings.ToLower(action.key)
	dataList, ok := redisBucket.dataMap[lowKey]
	if !ok {
		// 如果key不在，创建list，加入
		dataList = singlylinkedlist.New()
		dataList.Append(action.value)
		redisBucket.dataMap[lowKey] = dataList
	} else {
		// 同名的放在一起
		dataList.Append(action.value)
	}
	base.GLog.Debug("bucketid[%d] key[%s] set data:%+v", redisBucket.bucketID, lowKey, action.value)
}

func (redisBucket *RedisDataBucket) doDelete(action *DataActionInfoS) {
	// name is key
	lowKey := strings.ToLower(action.key)
	dataList, ok := redisBucket.dataMap[lowKey]
	if ok {
		if dataList.Size() == 1 {
			// 只有一个元素，直接删除map节点
			delete(redisBucket.dataMap, lowKey)
			base.GLog.Debug("bucketid[%d] key[%s] delete data", redisBucket.bucketID, lowKey)
		} else {
			// 找到确定的元素删除
			var pos int = -1
			dataList.Any(func(index int, value interface{}) bool {
				if dataMetaI, ok := value.(proto.DataMetaI); ok {
					if dataMetaI.Compare(action.value) {
						pos = index
						return true
					}
				}
				return false
			})
			base.GLog.Debug("bucketid[%d] key[%s] delete data pos[%d]", redisBucket.bucketID, lowKey, pos)
			if pos > -1 {
				dataList.Remove(pos)
			}
		}
	} else {
		base.GLog.Debug("bucketid[%d] key[%s] does not exist!", redisBucket.bucketID, lowKey)
	}
}
