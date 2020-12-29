/*
 * @Author: calmwu
 * @Date: 2017-10-23 15:01:03
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-11-09 12:53:21
 * @Comment:
 */
package redistool

import (
	"errors"
	"fmt"

	"github.com/fzzy/radix/redis"
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
	REDIS_HGET // 得到完整的map数据
	REDIS_HMGET
	REDIS_HSET // hash对象的完整更新
	REDIS_HDEL // 删除hash对象中的多个field
	REDIS_LGET // 得到完整的list数据
	REDIS_LSET // list的整体更新，把原有的key删除，重新设置
	REDIS_LRPUSH
	REDIS_LLPUSH
	REDIS_LREM  // 移除列表元素
	REDIS_LTRIM // 保留列表指定区间的元素
	REDIS_LLEN  // 获取列表长度
	REDIS_LPOP  // 移出并获取第一个元素
	REDIS_SGET  // set的整体获取
	REDIS_SSET  // set的整体更新，把原有的key删除，重新设置
	REDIS_PIPELINE
	REDIS_CLUSTERSLOTS
	REDIS_EXISTS // 判断key是否存在
	REDIS_SCRIPT_LOAD
	REDIS_EVAlSHA
	REDIS_INCR
	REDIS_ZINCRBY // 有序集合中对指定成员的分数加上增量 increment
	REDIS_ZSCORE  // 返回有序集中，成员的分数值
	REDIS_ZREM    // 移除有序集中的一个或多个成员
	REDIS_EXPIRE  // 设置Key生存时间
	REDIS_ZRRANK  // 返回倒序索引
)

const redisCmdName = "REDIS_GETREDIS_SETREDIS_HGETREDIS_HSETREDIS_LGETREDIS_LSETREDIS_LRPUSHREDIS_SGETREDIS_SSETREDIS_PIPELINE"

var redisCmdIndex = [...]uint8{0, 9, 18, 28, 38, 48, 58, 70, 81, 90, 104}

func (rc RedisCmd) String() string {
	if rc < 0 || rc+1 >= RedisCmd(len(redisCmdIndex)) {
		return fmt.Sprintf("RedisCmd(%d)", rc)
	}
	return redisCmdName[redisCmdIndex[rc]:redisCmdIndex[rc+1]]
}

// 内部返回类型
type RedisResultS struct {
	Ok     bool
	Result interface{}
}

type RedisCommandData struct {
	cmd       RedisCmd
	key       string
	value     interface{}
	ttl       int
	args      []interface{}
	replyChan chan<- *RedisResultS
}

type RedisPipeLineParamS struct {
	redisCmd      string
	containerType RedisContainerType
	args          []interface{}
}

type RedisPipeLineResultS struct {
	RedisCmd      string
	ContainerType RedisContainerType
	Key           string
	reply         *redis.Reply
}

type RedisPipeLineExecResult []*RedisPipeLineResultS

type RedisPipeLine struct {
	paramSlice []*RedisPipeLineParamS
}

type RedisClusterNodeS struct {
	IP   string
	Port int64
	Key  string
}

type RedisClusterSlotS struct {
	RedisSvrAddr string
	BeginPos     int64
	EndPos       int64
	NodeInfo     RedisClusterNodeS
}

func (rplR *RedisPipeLineResultS) Error() error {
	return rplR.reply.Err
}

func (rplR *RedisPipeLineResultS) String() (string, error) {
	//if rplR.ContainerType == REDIS_CONTAINER_STR {
	if rplR.reply.Type == redis.BulkReply {
		return rplR.reply.Str()
	}
	//}
	return "", errors.New("cmd is not get")
}

func (rplR *RedisPipeLineResultS) Hash() (map[string]string, error) {
	//if rplR.ContainerType == REDIS_CONTAINER_HASH {
	if rplR.reply.Type == redis.MultiReply {
		return rplR.reply.Hash()
	}
	//}
	// BulkReply
	return nil, errors.New("cmd is not get")
}

func (rplR *RedisPipeLineResultS) List() ([]string, error) {
	//if rplR.ContainerType == REDIS_CONTAINER_LIST {
	if rplR.reply.Type == redis.MultiReply {
		return rplR.reply.List()
	}
	//}
	return nil, errors.New("cmd is not get")
}
