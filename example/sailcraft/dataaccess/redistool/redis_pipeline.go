/*
 * @Author: calmwu
 * @Date: 2017-11-09 11:17:39
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-11-09 12:54:13
 */

package redistool

import (
	"fmt"

	"github.com/fzzy/radix/redis"
)

func redisPipeLine(conn *redis.Client, redisCmdData *RedisCommandData) (RedisPipeLineExecResult, error) {
	if redisPipeLineParams, ok := redisCmdData.value.([]*RedisPipeLineParamS); ok {
		redisPipeLineExecResult := make(RedisPipeLineExecResult, len(redisPipeLineParams))

		for index, _ := range redisPipeLineParams {

			result := &RedisPipeLineResultS{
				RedisCmd:      redisPipeLineParams[index].redisCmd,
				ContainerType: redisPipeLineParams[index].containerType,
				Key:           redisPipeLineParams[index].args[0].(string),
			}

			conn.Append(redisPipeLineParams[index].redisCmd, redisPipeLineParams[index].args...)

			redisPipeLineExecResult[index] = result
		}

		for index, _ := range redisPipeLineExecResult {
			redisPipeLineExecResult[index].reply = conn.GetReply()
		}

		return redisPipeLineExecResult, nil
	} else {
		return nil, fmt.Errorf("PipeLine value type is not []*RedisPipeLineParamsS")
	}
}
