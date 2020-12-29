/*
 * @Author: calmwu
 * @Date: 2017-10-27 15:02:57
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-10-27 17:01:41
 * @Comment:
 */
package redistool

import (
	"fmt"

	"github.com/fzzy/radix/redis"
)

const (
	HSCAN_BATCHCOUNT = 100
)

func redisHashSet(conn *redis.Client, redisCmdData *RedisCommandData) error {
	hKey := redisCmdData.key

	if hValue, ok := redisCmdData.value.(map[string]string); !ok {
		return fmt.Errorf("key[%s] value type is not map[string]string", hKey)
	} else {
		err := conn.Cmd("HMSET", hKey, hValue).Err
		return err
	}
}

func redisHashGet(conn *redis.Client, redisCmdData *RedisCommandData) (map[string]string, error) {
	hKey := redisCmdData.key

	exists, err := conn.Cmd("EXISTS", hKey).Int()
	if err != nil {
		return nil, err
	} else {
		if exists == 0 {
			return nil, nil
		} else {
			hashLen, err := conn.Cmd("HLEN", hKey).Int()
			if err != nil {
				return nil, err
			} else {
				hValue := make(map[string]string)
				if hashLen <= 0 {
					return hValue, nil
				} else {
					var cursorStart int64
					for hashLen > 0 {
						reply := conn.Cmd("HSCAN", hKey, cursorStart, "COUNT", HSCAN_BATCHCOUNT)
						if reply.Err != nil {
							return nil, reply.Err
						} else {
							cursorStart, _ = reply.Elems[0].Int64()
							hashVal, err := reply.Elems[1].Hash()
							if err != nil {
								return nil, err
							} else {
								for key, value := range hashVal {
									hValue[key] = value
								}
								hashLen -= len(hashVal)
							}
						}
					}
					return hValue, nil
				}
			}
		}
	}
}

func redisHashDel(conn *redis.Client, redisCmdData *RedisCommandData) (int, error) {
	hKey := redisCmdData.key
	hFields, ok := redisCmdData.value.([]string)
	if !ok {
		return -1, fmt.Errorf("fields type(%T) is not []string", redisCmdData.value)
	}

	reply := conn.Cmd("HDEL", hKey, hFields)

	if reply.Err != nil {
		return -1, reply.Err
	} else {
		return reply.Int()
	}
}

func redisHashGetFields(conn *redis.Client, redisCmdData *RedisCommandData) ([]string, error) {
	hKey := redisCmdData.key
	hFields, ok := redisCmdData.value.([]string)
	if !ok {
		return nil, fmt.Errorf("fields type(%T) is not []string", redisCmdData.value)
	}

	reply := conn.Cmd("HMGET", hKey, hFields)
	if reply.Err != nil {
		return nil, reply.Err
	} else {
		return reply.List()
	}
}
