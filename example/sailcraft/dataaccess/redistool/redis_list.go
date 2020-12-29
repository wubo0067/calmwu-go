/*
 * @Author: calmwu
 * @Date: 2017-10-26 16:21:48
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-10-28 16:16:18
 * @Comment:
 */

package redistool

import (
	"fmt"

	"github.com/fzzy/radix/redis"
)

const (
	LSCAN_BATCHCOUNT = 100
)

func redisListSet(conn *redis.Client, redisCmdData *RedisCommandData) error {
	lKey := redisCmdData.key

	if lValue, ok := redisCmdData.value.([]string); !ok {
		return fmt.Errorf("key[%s] value type is not []string", lKey)
	} else {
		lstLen, _ := conn.Cmd("LLEN", lKey).Int()
		// 在末尾追加
		for index, _ := range lValue {
			err := conn.Cmd("RPUSH", lKey, lValue[index]).Err
			if err != nil {
				fmt.Printf("index[%d] error[%s]\n", index, err.Error())
				return err
			}
		}
		// 设置数据长度
		if lstLen > 0 {
			return conn.Cmd("LTRIM", lKey, lstLen, -1).Err
		}
	}

	return nil
}

func redisListGet(conn *redis.Client, redisCmdData *RedisCommandData) ([]string, error) {
	lKey := redisCmdData.key

	// 获取key长度
	lstLen, err := conn.Cmd("LLEN", lKey).Int()
	if err != nil {
		return nil, err
	} else {
		lValue := make([]string, lstLen)

		cursorStart := 0
		for cursorStart < lstLen {
			cursorStop := cursorStart + LSCAN_BATCHCOUNT
			scanValue, err := conn.Cmd("LRANGE", lKey, cursorStart, cursorStop).List()
			if err != nil {
				return nil, err
			} else {
				for index, _ := range scanValue {
					lValue[cursorStart+index] = scanValue[index]
				}
			}
			cursorStart += len(scanValue)
		}
		return lValue, nil
	}
}

func redisListRPush(conn *redis.Client, redisCmdData *RedisCommandData) (int, error) {
	lKey := redisCmdData.key
	if lValue, ok := redisCmdData.value.([]string); !ok {
		return 0, fmt.Errorf("key[%s] value type is not string[]", lKey)
	} else {
		reply := conn.Cmd("RPUSH", lKey, lValue)
		return reply.Int()
	}
}

func redisListLPush(conn *redis.Client, redisCmdData *RedisCommandData) (int, error) {
	lKey := redisCmdData.key
	if lValue, ok := redisCmdData.value.([]string); !ok {
		return 0, fmt.Errorf("key[%s] value type is not string[]", lKey)
	} else {
		reply := conn.Cmd("LPUSH", lKey, lValue)
		return reply.Int()
	}
}

func redisListLTrim(conn *redis.Client, redisCmdData *RedisCommandData) error {
	reply := conn.Cmd("LTRIM", redisCmdData.key, redisCmdData.args)
	if reply.Err != nil {
		return reply.Err
	}

	return nil
}

func redisListRem(conn *redis.Client, redisCmdData *RedisCommandData) (int, error) {
	lKey := redisCmdData.key
	if lValue, ok := redisCmdData.value.(string); !ok {
		return 0, fmt.Errorf("key[%s] value type is not string", lKey)
	} else {
		return conn.Cmd("LREM", lKey, lValue, redisCmdData.args).Int()
	}
}

func redisListLen(conn *redis.Client, redisCmdData *RedisCommandData) (int, error) {
	lKey := redisCmdData.key
	return conn.Cmd("LLEN", lKey).Int()
}

func redisListLPop(conn *redis.Client, redisCmdData *RedisCommandData) (string, error) {
	lKey := redisCmdData.key

	lstLen, err := conn.Cmd("LLEN", lKey).Int()
	if err != nil {
		return "", err
	} else {
		if lstLen <= 0 {
			return "", nil
		} else {
			return conn.Cmd("LPOP", lKey).Str()
		}
	}
}
