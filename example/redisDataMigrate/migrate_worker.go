package main

import (
	"strings"
	"sync"
	"time"

	"github.com/fzzy/radix/redis"
	"github.com/gwenn/murmurhash3"

	"fmt"
)

const (
	E_MIGRATEINFOCHAN_SIZE = 100000
)

type MigrateErrCode int

const (
	E_MIGRATE_ERRCODE_FAILED MigrateErrCode = iota - 1
	E_MIGRATE_ERRCODE_OK
	E_MIGRATE_ERRCODE_SRCANDDEST_ADDR_SAME
	E_MIGRATE_ERRCODE_GETREDDISCONN_FAILED
)

type MigrateInfo struct {
	RedisKey      string // 迁移的key
	RedisKeyType  string // 迁移的key类型
	RedisKeyTTL   int    // 迁移key的ttl
	FromRedisAddr string // 迁移的redis源
	IsFinal       bool   // 最后的结束数据，worker收到后会退出
}

func (m *MigrateInfo) String() string {
	content := fmt.Sprintf("RedisKey[%s] RedisKeyType[%s] RedisKeyTTL[%d] FromRedisAddr[%s] IsFinal[%v]",
		m.RedisKey, m.RedisKeyType, m.RedisKeyTTL, m.FromRedisAddr, m.IsFinal)
	return content
}

type MigrateStatisticsInfo struct {
	MigrateInfo
	MigrateResult MigrateErrCode
	ToRedisAddr   string
}

func (m *MigrateStatisticsInfo) String() string {
	content := fmt.Sprintf("MigrateInfo[%s] MigrateErrCode[%d] ToRedisAddr[%s]",
		m.MigrateInfo.String(), m.MigrateResult, m.ToRedisAddr)
	return content
}

type MigrateWorkerMgr struct {
	MigrateInfoChan           chan *MigrateInfo           // 迁移数据通道
	MigrateStatisticsInfoChan chan *MigrateStatisticsInfo // 统计结果通道
	WorkWaitGroup             *sync.WaitGroup
}

var (
	gMigrateWorkMgr *MigrateWorkerMgr = nil
)

func init() {
	if gMigrateWorkMgr == nil {
		gMigrateWorkMgr = new(MigrateWorkerMgr)
		// 初始化数据
		gMigrateWorkMgr.MigrateInfoChan = make(chan *MigrateInfo, E_MIGRATEINFOCHAN_SIZE)
		gMigrateWorkMgr.MigrateStatisticsInfoChan = make(chan *MigrateStatisticsInfo, E_MIGRATEINFOCHAN_SIZE)
		gMigrateWorkMgr.WorkWaitGroup = new(sync.WaitGroup)
	}
}

func (w *MigrateWorkerMgr) StartWorks(config *Config) {

	// 启动worker
	var index int = 0
	for index < config.MigrateWorkerCount {
		w.WorkWaitGroup.Add(1)
		go MigrateWorkRoutine(config, w)
		index = index + 1
	}

	// 启动统计
	w.WorkWaitGroup.Add(1)
	go MigrateStatisticsWorkRouting(config.MigrateWorkerCount, w)
	return
}

func (w *MigrateWorkerMgr) WaitWorks() {
	gLog.Debug("Wait all migrate work exit!")
	w.WorkWaitGroup.Wait()
}

func (w *MigrateWorkerMgr) StopWorks(config *Config) {
	// 发送结束标志
	gLog.Debug("Now push final event to work")
	var index int = 0
	finalData := MigrateInfo{IsFinal: true}
	for index < config.MigrateWorkerCount {
		w.MigrateInfoChan <- &finalData
		index = index + 1
	}
}

func (w *MigrateWorkerMgr) DispatchMigrateInfo(migrateInfo *MigrateInfo) {
	w.MigrateInfoChan <- migrateInfo
}

func MigrateWorkRoutine(config *Config, workMgr *MigrateWorkerMgr) {

	gLog.Debug("MigrateWorkRoutine start running")

	defer workMgr.WorkWaitGroup.Done()
L:
	for {
		select {
		case migrateInfo, ok := <-workMgr.MigrateInfoChan:
			if ok {
				//gLog.Debug("MigrateInfo[%+v]", migrateInfo)
				if migrateInfo.IsFinal {
					statisticsFinalInfo := &MigrateStatisticsInfo{
						MigrateInfo: MigrateInfo{IsFinal: true},
					}
					workMgr.MigrateStatisticsInfoChan <- statisticsFinalInfo
					gLog.Debug("Receive Migrate FinalData, worker will Exit!")
					break L
				} else {
					//gLog.Debug("Process migrateInfo[%+v]", migrateInfo)

					// 处理完毕给统计发送数据
					statisticsInfo := new(MigrateStatisticsInfo)
					statisticsInfo.MigrateResult = 0

					// 处理数据
					toRedisAddr := CalcDestRedisSvrAddr(config, migrateInfo)
					// 如果源地址和目标地址相同，直接跳过
					if !strings.EqualFold(migrateInfo.FromRedisAddr, toRedisAddr) {
						if strings.Compare(migrateInfo.RedisKeyType, "string") == 0 {
							// 迁移string类型数据
							statisticsInfo.MigrateResult = MigrateStringData(toRedisAddr, migrateInfo)
						} else if strings.Compare(migrateInfo.RedisKeyType, "list") == 0 {
							// 迁移list类型数据
							statisticsInfo.MigrateResult = MigrateListData(toRedisAddr, migrateInfo, config)
						} else if strings.Compare(migrateInfo.RedisKeyType, "hash") == 0 {
							// hash
							statisticsInfo.MigrateResult = MigrateHashData(toRedisAddr, migrateInfo, config)
						} else if strings.Compare(migrateInfo.RedisKeyType, "set") == 0 {
							// set
							statisticsInfo.MigrateResult = MigrateSetData(toRedisAddr, migrateInfo, config)
						}
					} else {
						statisticsInfo.MigrateResult = 1
						gLog.Debug("Migrate source and destination addresses of the same address, [%s]",
							toRedisAddr)
					}

					statisticsInfo.FromRedisAddr = migrateInfo.FromRedisAddr
					statisticsInfo.IsFinal = migrateInfo.IsFinal
					statisticsInfo.RedisKey = migrateInfo.RedisKey
					statisticsInfo.ToRedisAddr = toRedisAddr

					statisticsInfo.RedisKeyType = migrateInfo.RedisKeyType
					workMgr.MigrateStatisticsInfoChan <- statisticsInfo
				}
			}
		}
	}
	gLog.Debug("MigrateWorkRoutine Exit!")
	return
}

func CalcDestRedisSvrAddr(config *Config, migrateInfo *MigrateInfo) (toRedisAddr string) {
	migrateToServCount := len(config.MigrateToServs)
	// 根据key计算hash值
	// https://github.com/luapower/murmurhash3/blob/master/murmurhash3.lua
	// https://github.com/gwenn/murmurhash3
	hashVal := murmurhash3.Murmur3A([]byte(migrateInfo.RedisKey), 0)
	toServIndex := hashVal % uint32(migrateToServCount)
	gLog.Debug("redisKey[%s] hashVal[%d] modVal[%d]", migrateInfo.RedisKey, hashVal, toServIndex)
	return config.MigrateToServs[toServIndex].RedisSvrAddr
}

func GetRedisConnByAddr(fromRedisServAddr, toRedisServAddr string) (fromRedisConn, toRedisConn *redis.Client) {
	fromRedisConn = gRedisConnPoolMgr.GetRedisConn(fromRedisServAddr)
	toRedisConn = gRedisConnPoolMgr.GetRedisConn(toRedisServAddr)
	return
}

func SetKeyTTL(toRedisConn *redis.Client, redisKey string, keyTTL int) {
	// 判断ttl
	if keyTTL > 0 {
		_, err := toRedisConn.Cmd("EXPIRE", redisKey, keyTTL).Int()
		if err != nil {
			gLog.Error("Expire cmd key[%s] toRedisAddr[%s] failed! error[%s]",
				redisKey, toRedisConn.Conn.RemoteAddr, err.Error())
		}
	}

	if keyTTL == 0 {
		gLog.Warn("Expire cmd key[%s] ttl is zero!", redisKey)
	}
}

// string迁移
func MigrateStringData(toRedisServAddr string, migrateInfo *MigrateInfo) MigrateErrCode {
	fromRedisConn, toRedisConn := GetRedisConnByAddr(migrateInfo.FromRedisAddr, toRedisServAddr)
	if fromRedisConn == nil || toRedisConn == nil {
		return E_MIGRATE_ERRCODE_GETREDDISCONN_FAILED
	}
	defer gRedisConnPoolMgr.PutRedisConn(migrateInfo.FromRedisAddr, fromRedisConn)
	defer gRedisConnPoolMgr.PutRedisConn(toRedisServAddr, toRedisConn)

	// 获取数据
	sValue, err := fromRedisConn.Cmd("GET", migrateInfo.RedisKey).Str()
	if err != nil {
		gLog.Error("Get string key[%s] type[%s] fromRedisAddr[%s] failed! error[%s]",
			migrateInfo.RedisKey, migrateInfo.RedisKeyType, migrateInfo.FromRedisAddr, err.Error())
		return E_MIGRATE_ERRCODE_FAILED
	}

	// 设置数据
	err = toRedisConn.Cmd("SET", migrateInfo.RedisKey, sValue).Err
	if err != nil {
		gLog.Error("SET cmd key[%s] type[%s] fromRedisAddr[%s] toRedisAddr[%s] failed! error[%s]",
			migrateInfo.RedisKey, migrateInfo.RedisKeyType, migrateInfo.FromRedisAddr, toRedisServAddr,
			err.Error())
		return E_MIGRATE_ERRCODE_FAILED
	}

	SetKeyTTL(toRedisConn, migrateInfo.RedisKey, migrateInfo.RedisKeyTTL)

	gLog.Info("Set Value key[%s] type[%s] fromRedisAddr[%s] toRedisAddr[%s] successed!",
		migrateInfo.RedisKey, migrateInfo.RedisKeyType, migrateInfo.FromRedisAddr, toRedisServAddr)
	return E_MIGRATE_ERRCODE_OK
}

// list数据迁移
func MigrateListData(toRedisServAddr string, migrateInfo *MigrateInfo, config *Config) MigrateErrCode {
	fromRedisConn, toRedisConn := GetRedisConnByAddr(migrateInfo.FromRedisAddr, toRedisServAddr)
	if fromRedisConn == nil || toRedisConn == nil {
		return E_MIGRATE_ERRCODE_GETREDDISCONN_FAILED
	}
	defer gRedisConnPoolMgr.PutRedisConn(migrateInfo.FromRedisAddr, fromRedisConn)
	defer gRedisConnPoolMgr.PutRedisConn(toRedisServAddr, toRedisConn)

	// 获取list的长度
	lstLen, err := fromRedisConn.Cmd("LLEN", migrateInfo.RedisKey).Int()
	if err != nil {
		gLog.Error("LLEN cmd key[%s] type[%s] fromRedisAddr[%s] toRedisAddr[%s] failed! error[%s]",
			migrateInfo.RedisKey, migrateInfo.RedisKeyType, migrateInfo.FromRedisAddr, toRedisServAddr,
			err.Error())
		return E_MIGRATE_ERRCODE_FAILED
	}

	gLog.Info("fromRedisAddr[%s] key[%s] lstLen[%d]", migrateInfo.FromRedisAddr, migrateInfo.RedisKey,
		lstLen)

	if lstLen <= 0 {
		// 如果是个空list，直接插入一个空字符串
		err = toRedisConn.Cmd("RPUSH", migrateInfo.RedisKey, "").Err
	} else {
		// 分批
		cursor_start := 0
		for cursor_start < lstLen {
			cursor_stop := cursor_start + config.LScanCount
			lValue, err := fromRedisConn.Cmd("LRANGE", migrateInfo.RedisKey, cursor_start, cursor_stop).List()
			if err != nil {
				gLog.Error("LRANGE cmd key[%s] cursor_start[%d] cursor_stop[%d] cursorfromRedisAddr[%s] toRedisAddr[%s] failed! error[%s]",
					migrateInfo.RedisKey, cursor_start, cursor_stop, migrateInfo.FromRedisAddr, toRedisServAddr,
					err.Error())
				return E_MIGRATE_ERRCODE_FAILED
			}
			cursor_start += len(lValue)

			// 写入
			err = toRedisConn.Cmd("RPUSH", migrateInfo.RedisKey, lValue).Err
			if err != nil {
				gLog.Error("RPUSH cmd key[%s] fromRedisAddr[%s] toRedisAddr[%s] failed! error[%s]",
					migrateInfo.RedisKey, migrateInfo.FromRedisAddr, toRedisServAddr, err.Error())
				return E_MIGRATE_ERRCODE_FAILED
			}
		}
	}

	// 获取list数据
	// lValue, err := fromRedisConn.Cmd("LRANGE", migrateInfo.RedisKey, 0, -1).List()
	// if err != nil {
	// 	gLog.Error("LRANGE cmd key[%s] type[%s] fromRedisAddr[%s] toRedisAddr[%s] failed! error[%s]",
	// 		migrateInfo.RedisKey, migrateInfo.RedisKeyType, migrateInfo.FromRedisAddr, toRedisServAddr,
	// 		err.Error())
	// 	return E_MIGRATE_ERRCODE_FAILED
	// }
	// //gLog.Debug("List key[%s] value[%+v]", migrateInfo.RedisKey, lValue)
	// // 设置数据
	// err = toRedisConn.Cmd("RPUSH", migrateInfo.RedisKey, lValue).Err
	// if err != nil {
	// 	gLog.Error("RPUSH cmd key[%s] type[%s] fromRedisAddr[%s] toRedisAddr[%s] failed! error[%s]",
	// 		migrateInfo.RedisKey, migrateInfo.RedisKeyType, migrateInfo.FromRedisAddr, toRedisServAddr,
	// 		err.Error())
	// 	return E_MIGRATE_ERRCODE_FAILED
	// }

	SetKeyTTL(toRedisConn, migrateInfo.RedisKey, migrateInfo.RedisKeyTTL)

	gLog.Info("RPUSH list key[%s] type[%s] fromRedisAddr[%s] toRedisAddr[%s] successed!",
		migrateInfo.RedisKey, migrateInfo.RedisKeyType, migrateInfo.FromRedisAddr, toRedisServAddr)

	return E_MIGRATE_ERRCODE_OK
}

// hash数据迁移
func MigrateHashData(toRedisServAddr string, migrateInfo *MigrateInfo, config *Config) MigrateErrCode {
	fromRedisConn, toRedisConn := GetRedisConnByAddr(migrateInfo.FromRedisAddr, toRedisServAddr)
	if fromRedisConn == nil || toRedisConn == nil {
		return E_MIGRATE_ERRCODE_GETREDDISCONN_FAILED
	}
	defer gRedisConnPoolMgr.PutRedisConn(migrateInfo.FromRedisAddr, fromRedisConn)
	defer gRedisConnPoolMgr.PutRedisConn(toRedisServAddr, toRedisConn)

	// 得到hash数据长度
	hashLen, err := fromRedisConn.Cmd("HLEN", migrateInfo.RedisKey).Int()
	if err != nil {
		gLog.Error("HLEN cmd key[%s] fromRedisAddr[%s] toRedisAddr[%s] failed! error[%s]",
			migrateInfo.RedisKey, migrateInfo.FromRedisAddr, toRedisServAddr, err.Error())
		return E_MIGRATE_ERRCODE_FAILED
	}

	gLog.Debug("fromRedisAddr[%s] key[%s] hashLen[%d]", migrateInfo.FromRedisAddr, migrateInfo.RedisKey,
		hashLen)

	if hashLen <= 0 {
		// 设置一个空集
		err = fromRedisConn.Cmd("HSET", migrateInfo.RedisKey, "", "").Err
	} else {
		var scanCursor int64 = 0
		for hashLen > 0 {
			reply := fromRedisConn.Cmd("HSCAN", migrateInfo.RedisKey, scanCursor, "COUNT", config.HScanCount)
			if reply.Err != nil {
				gLog.Error("HSCAN cmd key[%s] fromRedisAddr[%s] toRedisAddr[%s] failed! error[%s]",
					migrateInfo.RedisKey, migrateInfo.FromRedisAddr, toRedisServAddr, err.Error())
				return E_MIGRATE_ERRCODE_FAILED
			}

			scanCursor, _ = reply.Elems[0].Int64()
			hashVal, err := reply.Elems[1].Hash()
			if err != nil {
				gLog.Error("HSCAN get HashVal key[%s] fromRedisAddr[%s] toRedisAddr[%s] failed! error[%s]",
					migrateInfo.RedisKey, migrateInfo.FromRedisAddr, toRedisServAddr, err.Error())
				return E_MIGRATE_ERRCODE_FAILED
			}

			// 设置数据
			err = toRedisConn.Cmd("HMSET", migrateInfo.RedisKey, hashVal).Err
			if err != nil {
				gLog.Error("HMSET cmd key[%s] type[%s] fromRedisAddr[%s] toRedisAddr[%s] failed! error[%s]",
					migrateInfo.RedisKey, migrateInfo.RedisKeyType, migrateInfo.FromRedisAddr, toRedisServAddr,
					err.Error())
				return E_MIGRATE_ERRCODE_FAILED
			}

			hashLen -= len(hashVal)
			gLog.Debug("migrated key[%s] hash data count[%d], remaining set data count[%d]",
				migrateInfo.RedisKey, len(hashVal), hashLen)
		}
	}

	// 获取hash数据
	// hValue, err := fromRedisConn.Cmd("HGETALL", migrateInfo.RedisKey).Hash()
	// if err != nil {
	// 	gLog.Error("HGETALL cmd key[%s] fromRedisAddr[%s] toRedisAddr[%s] failed! error[%s]",
	// 		migrateInfo.RedisKey, migrateInfo.FromRedisAddr, toRedisServAddr,
	// 		err.Error())
	// 	return E_MIGRATE_ERRCODE_FAILED
	// }
	// //gLog.Debug("Hash key[%s] value[%+v]", migrateInfo.RedisKey, hValue)
	// // 设置数据
	// err = toRedisConn.Cmd("HMSET", migrateInfo.RedisKey, hValue).Err
	// if err != nil {
	// 	gLog.Error("HMSET cmd key[%s] type[%s] fromRedisAddr[%s] toRedisAddr[%s] failed! error[%s]",
	// 		migrateInfo.RedisKey, migrateInfo.RedisKeyType, migrateInfo.FromRedisAddr, toRedisServAddr,
	// 		err.Error())
	// 	return E_MIGRATE_ERRCODE_FAILED
	// }

	SetKeyTTL(toRedisConn, migrateInfo.RedisKey, migrateInfo.RedisKeyTTL)

	gLog.Info("HMSET hash key[%s] type[%s] fromRedisAddr[%s] toRedisAddr[%s] successed!",
		migrateInfo.RedisKey, migrateInfo.RedisKeyType, migrateInfo.FromRedisAddr, toRedisServAddr)

	return E_MIGRATE_ERRCODE_OK
}

// 迁移set数据
func MigrateSetData(toRedisServAddr string, migrateInfo *MigrateInfo, config *Config) MigrateErrCode {
	fromRedisConn, toRedisConn := GetRedisConnByAddr(migrateInfo.FromRedisAddr, toRedisServAddr)
	if fromRedisConn == nil || toRedisConn == nil {
		return E_MIGRATE_ERRCODE_GETREDDISCONN_FAILED
	}
	defer gRedisConnPoolMgr.PutRedisConn(migrateInfo.FromRedisAddr, fromRedisConn)
	defer gRedisConnPoolMgr.PutRedisConn(toRedisServAddr, toRedisConn)

	setLen, err := fromRedisConn.Cmd("SCARD", migrateInfo.RedisKey).Int()
	if err != nil {
		gLog.Error("SCARD cmd key[%s] fromRedisAddr[%s] toRedisAddr[%s] failed! error[%s]",
			migrateInfo.RedisKey, migrateInfo.FromRedisAddr, toRedisServAddr, err.Error())
		return E_MIGRATE_ERRCODE_FAILED
	}

	gLog.Debug("fromRedisAddr[%s] key[%s] setLen[%d]", migrateInfo.FromRedisAddr, migrateInfo.RedisKey,
		setLen)

	if setLen <= 0 {
		err = toRedisConn.Cmd("SADD", migrateInfo.RedisKey, "").Err
	} else {
		var scanCursor int64 = 0
		for setLen > 0 {
			reply := fromRedisConn.Cmd("SSCAN", migrateInfo.RedisKey, scanCursor, "COUNT", config.SScanCount)
			if reply.Err != nil {
				gLog.Error("SSCAN cmd key[%s] fromRedisAddr[%s] toRedisAddr[%s] failed! error[%s]",
					migrateInfo.RedisKey, migrateInfo.FromRedisAddr, toRedisServAddr, err.Error())
				return E_MIGRATE_ERRCODE_FAILED
			}

			scanCursor, _ = reply.Elems[0].Int64()
			setVal, err := reply.Elems[1].List()
			if err != nil {
				gLog.Error("SSCAN get HashVal key[%s] fromRedisAddr[%s] toRedisAddr[%s] failed! error[%s]",
					migrateInfo.RedisKey, migrateInfo.FromRedisAddr, toRedisServAddr, err.Error())
				return E_MIGRATE_ERRCODE_FAILED
			}

			// 设置数据
			err = toRedisConn.Cmd("SADD", migrateInfo.RedisKey, setVal).Err
			if err != nil {
				gLog.Error("SADD cmd key[%s] type[%s] fromRedisAddr[%s] toRedisAddr[%s] failed! error[%s]",
					migrateInfo.RedisKey, migrateInfo.RedisKeyType, migrateInfo.FromRedisAddr, toRedisServAddr,
					err.Error())
				return E_MIGRATE_ERRCODE_FAILED
			}

			setLen -= len(setVal)
			gLog.Debug("migrated key[%s] hash data count[%d], remaining hash data count[%d]",
				migrateInfo.RedisKey, len(setVal), setLen)
		}
	}

	// // 获取set数据，如果set数据重多会引发redis阻塞
	// sValue, err := fromRedisConn.Cmd("SMEMBERS", migrateInfo.RedisKey).List()
	// if err != nil {
	// 	gLog.Error("SMEMBERS cmd key[%s] type[%s] fromRedisAddr[%s] toRedisAddr[%s] failed! error[%s]",
	// 		migrateInfo.RedisKey, migrateInfo.RedisKeyType, migrateInfo.FromRedisAddr, toRedisServAddr,
	// 		err.Error())
	// 	return E_MIGRATE_ERRCODE_FAILED
	// }
	// //gLog.Debug("Set key[%s] value[%+v]", migrateInfo.RedisKey, sValue)
	// // 设置数据
	// err = toRedisConn.Cmd("SADD", migrateInfo.RedisKey, sValue).Err
	// if err != nil {
	// 	gLog.Error("SADD cmd key[%s] type[%s] fromRedisAddr[%s] toRedisAddr[%s] failed! error[%s]",
	// 		migrateInfo.RedisKey, migrateInfo.RedisKeyType, migrateInfo.FromRedisAddr, toRedisServAddr,
	// 		err.Error())
	// 	return E_MIGRATE_ERRCODE_FAILED
	// }

	gLog.Info("SADD set key[%s] type[%s] fromRedisAddr[%s] toRedisAddr[%s] successed!",
		migrateInfo.RedisKey, migrateInfo.RedisKeyType, migrateInfo.FromRedisAddr, toRedisServAddr)

	return E_MIGRATE_ERRCODE_OK
}

func MigrateStatisticsWorkRouting(workCount int, workMgr *MigrateWorkerMgr) {

	gLog.Debug("MigrateStatisticsWorkRouting start running")
	defer workMgr.WorkWaitGroup.Done()

	statisticsRedisKeyCount := 0
	statisticsTicker := time.NewTicker(time.Second * 3)
	// 写统计文件，迁移的成功和失败
L:
	for {
		select {
		case migrateStatisticsInfo, ok := <-workMgr.MigrateStatisticsInfoChan:
			if ok {
				gLog.Debug("MigrateStatisticsInfo[%+v]", migrateStatisticsInfo)
				if migrateStatisticsInfo.IsFinal {
					workCount = workCount - 1
					gLog.Debug("Migrate Statistics work count[%d]", workCount)
					if workCount == 0 {
						gLog.Debug("Migrate Statistics worker will Exit! Total Process RedisKey count[%d]", statisticsRedisKeyCount)
						break L
					}
				} else {
					statisticsRedisKeyCount++
				}
			}
		case <-statisticsTicker.C:
			gLog.Debug("Now Process RedisKey count[%d]", statisticsRedisKeyCount)
		}
	}
	gLog.Debug("MigrateStatisticsWorkRouting Exit!")
	return
}
