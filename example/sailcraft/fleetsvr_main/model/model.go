package model

import (
	"sailcraft/dataaccess/mysql"
	"sailcraft/dataaccess/redistool"
	"sailcraft/fleetsvr_main/custom_errors"
)

const (
	PLATFORM_SET            = "sailcraft_platform_set"
	UIN_ALLOCATION_SET1     = "sailcraft_uin_allocation_set_1"
	UIN_ALLOCATION_SET2     = "sailcraft_uin_allocation_set_2"
	UIN_ALLOCATION_SET3     = "sailcraft_uin_allocation_set_3"
	UIN_SET_1_1000000       = "sailcraft_uin_set_1_1000000"
	UIN_SET_1000001_2000000 = "sailcraft_uin_set_1000001_2000000"
	UIN_SET_2000001_3000000 = "sailcraft_uin_set_2000001_3000000"
	UIN_SET_3000001_4000000 = "sailcraft_uin_set_3000001_4000000"
	UIN_SET_4000001_5000000 = "sailcraft_uin_set_4000001_5000000"

	REDIS_CONTAINER_STR = iota
	REDIS_CONTAINER_LIST
	REDIS_CONTAINER_SET
	REDIS_CONTAINER_HASH
)

type RedisCommand struct {
	Cmd           string
	Key           string
	ContainerType int
	Args          []interface{}
	Index         int
}

type RedisPipelineData struct {
	RedisMgr *redistool.RedisNode
	Commands []*RedisCommand
	Results  redistool.RedisPipeLineExecResult
}

type RedisKeyGroup struct {
	RedisMgr *redistool.RedisNode
	Keys     []string
}

func GetTableSplitIndex(uin int) int {
	return uin % 10
}

func GetUinSetDataBaseName(uin int) string {
	if uin > 0 && uin <= 1000000 {
		return UIN_SET_1_1000000
	} else if uin > 1000000 && uin <= 2000000 {
		return UIN_SET_1000001_2000000
	} else if uin > 2000000 && uin <= 3000000 {
		return UIN_SET_2000001_3000000
	} else if uin > 3000000 && uin <= 4000000 {
		return UIN_SET_3000001_4000000
	} else if uin > 4000000 && uin <= 5000000 {
		return UIN_SET_4000001_5000000
	} else {
		return ""
	}
}

func GetUinSetMysql(uin int) *mysql.DBEngineInfoS {
	dbName := GetUinSetDataBaseName(uin)
	if dbName != "" {
		engine, err := mysql.GMysqlManager.GetMysql(dbName)
		if err == nil && engine != nil {
			return engine
		}
	}

	return nil
}

func GetClusterRedis(key string) *redistool.RedisNode {
	if key != "" {
		redisMgr, err := redistool.GRedisManager.GetClusterRedisMgr(key)
		if err == nil {
			return redisMgr
		}
	}

	return nil
}

func GetSingletonRedis() *redistool.RedisNode {
	redisMgr, err := redistool.GRedisManager.GetSingletonRedisMgr()
	if err == nil {
		return redisMgr
	}

	return nil
}

func GetKeysGroupByClusterRedis(keys ...string) (map[string]*RedisKeyGroup, error) {
	redisMap := make(map[string]*RedisKeyGroup)
	for _, key := range keys {
		redisMgr := GetClusterRedis(key)
		if redisMgr == nil {
			return nil, custom_errors.New("redisMgr[%s] is empty", key)
		}

		keyGroup, ok := redisMap[redisMgr.RedisSvrAddr]

		if !ok {
			keyGroup = new(RedisKeyGroup)
			keyGroup.RedisMgr = redisMgr
			redisMap[redisMgr.RedisSvrAddr] = keyGroup
		}

		keyGroup.Keys = append(keyGroup.Keys, key)
	}

	return redisMap, nil
}

func ClusterRedisPipeline(commands ...*RedisCommand) ([]interface{}, error) {
	if len(commands) <= 0 {
		return nil, custom_errors.New("commands is empty")
	}

	result := make([]interface{}, len(commands))
	redisMap := make(map[string]*RedisPipelineData)
	for index, cmd := range commands {
		if cmd != nil {
			redisMgr := GetClusterRedis(cmd.Key)
			if redisMgr == nil {
				return nil, custom_errors.New("redisMgr[%s] is empty", cmd.Key)
			}

			cmd.Index = index
			redisPipelineData, ok := redisMap[redisMgr.RedisSvrAddr]
			if !ok {
				redisPipelineData = new(RedisPipelineData)
				redisPipelineData.RedisMgr = redisMgr
				redisMap[redisMgr.RedisSvrAddr] = redisPipelineData
			}
			redisPipelineData.Commands = append(redisPipelineData.Commands, cmd)
		} else {
			result[index] = nil
		}
	}

	for _, pipelineData := range redisMap {
		if len(pipelineData.Commands) > 0 {
			redisPipeline := redistool.NewRedisPipeLine()
			for _, cmd := range pipelineData.Commands {
				switch cmd.ContainerType {
				case REDIS_CONTAINER_STR:
					redisPipeline.Append(redistool.REDIS_CONTAINER_STR, cmd.Cmd, cmd.Key, cmd.Args)
				case REDIS_CONTAINER_SET:
					redisPipeline.Append(redistool.REDIS_CONTAINER_SET, cmd.Cmd, cmd.Key, cmd.Args)
				case REDIS_CONTAINER_LIST:
					redisPipeline.Append(redistool.REDIS_CONTAINER_LIST, cmd.Cmd, cmd.Key, cmd.Args)
				case REDIS_CONTAINER_HASH:
					redisPipeline.Append(redistool.REDIS_CONTAINER_HASH, cmd.Cmd, cmd.Key, cmd.Args)
				default:
					continue
				}
			}

			res, err := redisPipeline.Run(pipelineData.RedisMgr)
			if err != nil {
				return nil, err
			}

			pipelineData.Results = res

			for index, res := range pipelineData.Results {
				switch res.ContainerType {
				case REDIS_CONTAINER_STR:
					str, err := res.String()
					if err != nil {
						return nil, err
					}
					result[pipelineData.Commands[index].Index] = str
				case REDIS_CONTAINER_SET:
					set, err := res.List()
					if err != nil {
						return nil, err
					}
					result[pipelineData.Commands[index].Index] = set
				case REDIS_CONTAINER_LIST:
					list, err := res.List()
					if err != nil {
						return nil, err
					}
					result[pipelineData.Commands[index].Index] = list
				case REDIS_CONTAINER_HASH:
					hash, err := res.Hash()
					if err != nil {
						return nil, err
					}
					result[pipelineData.Commands[index].Index] = hash
				default:
					continue
				}
			}
		}
	}

	return result, nil
}
