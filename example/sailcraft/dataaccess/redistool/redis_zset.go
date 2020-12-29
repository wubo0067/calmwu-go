package redistool

import (
	"github.com/fzzy/radix/redis"
)

func redisZSetIncreBy(conn *redis.Client, redisCmdData *RedisCommandData) (int, error) {
	reply := conn.Cmd("ZINCRBY", redisCmdData.key, redisCmdData.args)
	return reply.Int()
}

func redisZSetScore(conn *redis.Client, redisCmdData *RedisCommandData) (int, error) {
	reply := conn.Cmd("ZSCORE", redisCmdData.key, redisCmdData.args)
	if reply.Type == redis.NilReply {
		return 0, nil
	} else {
		return reply.Int()
	}
}

func redisZsetRem(conn *redis.Client, redisCmdData *RedisCommandData) (int, error) {
	reply := conn.Cmd("ZREM", redisCmdData.key, redisCmdData.value)
	return reply.Int()
}

func redisZRevRank(conn *redis.Client, redisCmdData *RedisCommandData) (int, error) {
	reply := conn.Cmd("ZREVRANK", redisCmdData.key, redisCmdData.value)
	if reply.Type == redis.NilReply {
		return -1, nil
	}

	return reply.Int()
}
