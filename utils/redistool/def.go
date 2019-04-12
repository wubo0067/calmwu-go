/*
 * @Author: calmwu
 * @Date: 2017-10-23 15:01:03
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-30 19:37:43
 * @Comment:
 */
package redistool

import (
	"github.com/go-redis/redis"
	"golang.org/x/sys/cpu"
)

// 命令类型
type RedisCmd int
type RedisContainerType int

const (
	REDIS_CONTAINER_STR = iota
	REDIS_CONTAINER_LIST
	REDIS_CONTAINER_SET
	REDIS_CONTAINER_HASH
)

// 命令字
const (
	REDIS_GET RedisCmd = iota
	REDIS_SET
	REDIS_SETNX
	REDIS_DEL
	REDIS_HGETALL // 得到完整的map数据
	REDIS_HMGET
	REDIS_HSET
	REDIS_HMSET
	REDIS_HDEL // 删除hash对象中的多个field
	REDIS_LGET // 得到完整的list数据
	REDIS_LSET // list的整体更新，把原有的key删除，重新设置
	REDIS_LRPUSH
	REDIS_LLPUSH
	REDIS_LREM      // 移除列表元素
	REDIS_LTRIM     // 保留列表指定区间的元素
	REDIS_LLEN      // 获取列表长度
	REDIS_LPOP      // 移出并获取第一个元素
	REDIS_SGET      // set的整体获取
	REDIS_SSET      // set的整体更新，把原有的key删除，重新设置
	REDIS_SISMEMBER // 判断是否再set中
	REDIS_SADD      // set add
	REDIS_SREM      // set remove
	REDIS_PIPELINE
	REDIS_CLUSTERSLOTS
	REDIS_EXISTS // 判断key是否存在
	REDIS_SCRIPT_LOAD
	REDIS_EVAlSHA
	REDIS_INCR
	REDIS_ZINCRBY     // 有序集合中对指定成员的分数加上增量 increment
	REDIS_ZSCORE      // 返回有序集中，成员的分数值
	REDIS_ZREM        // 移除有序集中的一个或多个成员
	REDIS_EXPIRE      // 设置Key生存时间
	REDIS_ZRRANK      // 返回倒序索引
	REDIS_SCAN        // key扫描
	REDIS_CLUSTERSCAN // cluster scan
	REDIS_DBSIZE      // key的数量
)

// 内部返回类型
type RedisResultS struct {
	Ok     bool
	Result interface{}
}

type RedisScanResult struct {
	keys   []string
	cursor uint64
}

type RedisCommandData struct {
	_         cpu.CacheLinePad
	cmd       RedisCmd
	key       string
	fieldKey  string
	value     interface{}
	ttl       int
	args      []interface{}
	replyChan chan<- *RedisResultS
	_         cpu.CacheLinePad
}

type redisPipelineFn func(redis.Pipeliner) error
