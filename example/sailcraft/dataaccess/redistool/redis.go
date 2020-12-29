/*
 * @Author: calmwu
 * @Date: 2017-10-23 14:59:25
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-11-10 15:42:20
 * @Comment:
 */

package redistool

import (
	"fmt"
	"sailcraft/base"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fzzy/radix/extra/pool"
	"github.com/fzzy/radix/redis"
)

type RedisNode struct {
	RedisSvrAddr   string
	sessionCount   int
	redisConnPool  *pool.Pool
	redisCmdChan   chan *RedisCommandData
	redisExitChan  chan struct{}
	redisRoutineWG *sync.WaitGroup
}

func NewRedis(RedisSvrAddr string, sessionCount int) *RedisNode {
	redisMgr := new(RedisNode)
	redisMgr.RedisSvrAddr = RedisSvrAddr
	redisMgr.sessionCount = sessionCount
	return redisMgr
}

func (rn *RedisNode) Start() error {
	if rn.redisConnPool == nil {
		rn.redisCmdChan = make(chan *RedisCommandData, 1000)
		rn.redisExitChan = make(chan struct{})
		rn.redisRoutineWG = new(sync.WaitGroup)

		base.GLog.Debug("Redis host[%s]", rn.RedisSvrAddr)

		df := func(network, addr string) (*redis.Client, error) {
			client, err := redis.DialTimeout(network, addr, time.Duration(time.Second))
			if err != nil {
				base.GLog.Error("Connect to Redis[%s] failed! error[%s]", addr, err.Error())
				return nil, err
			}
			return client, nil
		}
		connPool, err := pool.NewCustomPool("tcp", rn.RedisSvrAddr, rn.sessionCount, df)
		if err != nil {
			return err
		} else {
			rn.redisConnPool = connPool
			var index = 0
			for index < rn.sessionCount {
				// 启动命令处理协程
				go redisCmdRun(rn)
				rn.redisRoutineWG.Add(1)
				index++
			}
		}
		return nil
	}
	return fmt.Errorf("RedisConnPool already init")
}

func (rn *RedisNode) Stop() {
	close(rn.redisExitChan)
	rn.redisRoutineWG.Wait()
	rn.redisConnPool.Empty()
}

func (rn *RedisNode) Addr() string {
	return rn.RedisSvrAddr
}

func redisCmdRun(rn *RedisNode) {
	base.GLog.Debug("redis cmd process routine running")
	defer rn.redisRoutineWG.Done()

L:
	for {
		select {
		case redisCmdData, ok := <-rn.redisCmdChan:
			if ok {
				conn, err := rn.redisConnPool.Get()
				if err == nil {
					switch redisCmdData.cmd {
					case REDIS_GET:
						// 通过key获得value，以[]byte形式返回
						reply := conn.Cmd("GET", redisCmdData.key)
						// base.GLog.Debug("Reply Type:%d Elems:[%+v] Error:%v", reply.Type, reply.Elems, reply.Err)
						if reply.Type == redis.NilReply {
							redisCmdData.replyChan <- &RedisResultS{Ok: false, Result: ErrKeyNotExist(redisCmdData.key)}
							rn.redisConnPool.Put(conn)
						} else {
							sValue, err := reply.Bytes()
							if err == nil {
								redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: sValue}
								rn.redisConnPool.Put(conn)
							}
						}

					case REDIS_SET:
						// 设置key value
						err = conn.Cmd("SET", redisCmdData.key, redisCmdData.value).Err
						if err == nil {
							// set成功
							redisCmdData.replyChan <- &RedisResultS{Ok: true}
							rn.redisConnPool.Put(conn)
						}
					case REDIS_SETNX:
						// 设置key value
						reply := conn.Cmd("SET", redisCmdData.key, redisCmdData.value, "EX", redisCmdData.ttl, "NX")
						if reply.Type == redis.StatusReply {
							// set成功
							redisCmdData.replyChan <- &RedisResultS{Ok: true}
							rn.redisConnPool.Put(conn)
						}
					case REDIS_DEL:
						// 设置key value
						err = conn.Cmd("DEL", redisCmdData.key).Err
						if err == nil {
							// set成功
							redisCmdData.replyChan <- &RedisResultS{Ok: true}
							rn.redisConnPool.Put(conn)
						}
					case REDIS_LSET:
						err = redisListSet(conn, redisCmdData)
						if err == nil {
							// set成功
							redisCmdData.replyChan <- &RedisResultS{Ok: true}
							rn.redisConnPool.Put(conn)
						}
					case REDIS_LGET:
						var lValue []string
						lValue, err = redisListGet(conn, redisCmdData)
						if err == nil {
							redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: lValue}
							rn.redisConnPool.Put(conn)
						}
					case REDIS_LRPUSH:
						sValue, err := redisListRPush(conn, redisCmdData)
						if err == nil {
							// set成功
							redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: sValue}
							rn.redisConnPool.Put(conn)
						}
					case REDIS_LLPUSH:
						sValue, err := redisListLPush(conn, redisCmdData)
						if err == nil {
							redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: sValue}
							rn.redisConnPool.Put(conn)
						}
					case REDIS_LREM:
						delCount, err := redisListRem(conn, redisCmdData)
						if err == nil {
							redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: delCount}
							rn.redisConnPool.Put(conn)
						}
					case REDIS_LTRIM:
						err := redisListLTrim(conn, redisCmdData)
						if err == nil {
							redisCmdData.replyChan <- &RedisResultS{Ok: true}
						}
					case REDIS_LLEN:
						len, err := redisListLen(conn, redisCmdData)
						if err == nil {
							redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: len}
							rn.redisConnPool.Put(conn)
						}
					case REDIS_LPOP:
						str, err := redisListLPop(conn, redisCmdData)
						if err == nil {
							redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: str}
							rn.redisConnPool.Put(conn)
						}
					case REDIS_HGET:
						var hValue map[string]string
						hValue, err = redisHashGet(conn, redisCmdData)
						if err == nil {
							redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: hValue}
							rn.redisConnPool.Put(conn)
						}
					case REDIS_HMGET:
						var hValue []string
						hValue, err = redisHashGetFields(conn, redisCmdData)
						if err == nil {
							redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: hValue}
							rn.redisConnPool.Put(conn)
						}
					case REDIS_HSET:
						err = redisHashSet(conn, redisCmdData)
						if err == nil {
							// set成功
							redisCmdData.replyChan <- &RedisResultS{Ok: true}
							rn.redisConnPool.Put(conn)
						}
					case REDIS_HDEL:
						delCount, err := redisHashDel(conn, redisCmdData)
						if err == nil {
							// 删除成功
							redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: delCount}
							rn.redisConnPool.Put(conn)
						}
					case REDIS_PIPELINE:
						var plRes RedisPipeLineExecResult
						plRes, err = redisPipeLine(conn, redisCmdData)
						if err == nil {
							redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: plRes}
							rn.redisConnPool.Put(conn)
						}
					case REDIS_CLUSTERSLOTS:
						var csValue []*RedisClusterSlotS
						csValue, err = redisClusterSlotsGet(conn, redisCmdData)
						if err == nil {
							redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: csValue}
							rn.redisConnPool.Put(conn)
						}
						// case REDIS_SGET:
						// 	sValue, err := redisSetGet(conn, redisCmdData)
						// 	if err == nil {
						// 		redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: sValue}
						// 		rn.redisConnPool.Put(conn)
						// 	}
						// case REDIS_SSET:
						// 	err := redisSetSet(conn, redisCmdData)
						// 	if err == nil {
						// 		// set成功
						// 		redisCmdData.replyChan <- &RedisResultS{Ok: true}
						// 		rn.redisConnPool.Put(conn)
						// 	}
					case REDIS_EXISTS:
						var bExists bool = false
						bExists, err = conn.Cmd("EXISTS", redisCmdData.key).Bool()
						if err == nil {
							redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: bExists}
							rn.redisConnPool.Put(conn)
						}
					case REDIS_SCRIPT_LOAD:
						reply := conn.Cmd("script", "load", redisCmdData.value)
						//base.GLog.Debug("REDIS_SCRIPT_LOAD Reply Type: %d, Elems: %+v, Err: %+v", reply.Type, reply.Elems, reply.Err)
						sValue, err := reply.Bytes()
						if err == nil {
							redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: sValue}
							rn.redisConnPool.Put(conn)
						}
					case REDIS_EVAlSHA:
						reply := conn.Cmd("EVALSHA", redisCmdData.args...)
						//base.GLog.Debug("REDIS_EVAlSHA Reply Type: %d, Elems: %+v, Err: %+v", reply.Type, reply.Elems, reply.Err)
						sValue, err := reply.Int()
						if err == nil {
							redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: sValue}
							rn.redisConnPool.Put(conn)
						}

					case REDIS_INCR:
						reply := conn.Cmd("INCR", redisCmdData.key, redisCmdData.args)
						sValue, err := reply.Int()
						if err == nil {
							redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: sValue}
							rn.redisConnPool.Put(conn)
						}
					case REDIS_ZINCRBY:
						sValue, err := redisZSetIncreBy(conn, redisCmdData)
						if err == nil {
							redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: sValue}
							rn.redisConnPool.Put(conn)
						}
					case REDIS_ZSCORE:
						sValue, err := redisZSetScore(conn, redisCmdData)
						if err == nil {
							redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: sValue}
							rn.redisConnPool.Put(conn)
						}
					case REDIS_ZREM:
						sValue, err := redisZsetRem(conn, redisCmdData)
						if err == nil {
							redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: sValue}
							rn.redisConnPool.Put(conn)
						}
					case REDIS_ZRRANK:
						sValue, err := redisZRevRank(conn, redisCmdData)
						if err == nil {
							redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: sValue}
							rn.redisConnPool.Put(conn)
						}
					case REDIS_EXPIRE:
						reply := conn.Cmd("EXPIRE", redisCmdData.key, redisCmdData.value)
						sValue, err := reply.Int()
						if err == nil {
							redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: sValue}
							rn.redisConnPool.Put(conn)
						}
					}

					if err != nil {
						base.GLog.Error("redis[%s] key[%s]! reason[%s]", conn.Conn.RemoteAddr().String(), redisCmdData.key, err.Error())
						redisCmdData.replyChan <- &RedisResultS{Ok: false, Result: err}
						if strings.Contains(err.Error(), "use of closed network connection") {
							// 连接断开了
							conn.Close()
						} else {
							rn.redisConnPool.Put(conn)
						}
					}
				} else {
					base.GLog.Error("get redis conn failed! reason[%s]", err.Error())
					result := &RedisResultS{Ok: false, Result: nil}
					redisCmdData.replyChan <- result
				}
			}
		case <-rn.redisExitChan:
			break L
		}
	}
	base.GLog.Warn("redisCmdRun Exit")
	return
}

func (rn *RedisNode) waitResult(reply <-chan *RedisResultS, key string) (interface{}, error) {
	select {
	case redisRes, ok := <-reply:
		if ok {
			if redisRes.Ok {
				return redisRes.Result, nil
			} else {
				return nil, redisRes.Result.(error)
			}
		}
	case <-time.After(2 * time.Second):
		return nil, ErrTimeOut(key)
	}
	return nil, ErrTimeOut(key)
}

func (rn *RedisNode) StringGet(key string) (interface{}, error) {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_GET,
		key:       key,
		replyChan: reply,
	}
	return rn.waitResult(reply, key)
}

func (rn *RedisNode) StringSet(key string, value []byte) error {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_SET,
		key:       key,
		value:     value,
		replyChan: reply,
	}
	_, err := rn.waitResult(reply, key)
	return err
}

func (rn *RedisNode) StringSetNX(key string, value []byte, ttl int) error {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_SETNX,
		key:       key,
		value:     value,
		ttl:       ttl,
		replyChan: reply,
	}
	_, err := rn.waitResult(reply, key)
	return err
}

func (rn *RedisNode) DelKey(key string) error {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_DEL,
		key:       key,
		replyChan: reply,
	}
	_, err := rn.waitResult(reply, key)
	return err
}

func (rn *RedisNode) ListSet(key string, value []string) error {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_LSET,
		key:       key,
		value:     value,
		replyChan: reply,
	}
	_, err := rn.waitResult(reply, key)
	return err
}

func (rn *RedisNode) ListGet(key string) ([]string, error) {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_LGET,
		key:       key,
		replyChan: reply,
	}
	lValue, err := rn.waitResult(reply, key)
	if err == nil {
		return lValue.([]string), nil
	}
	return nil, err
}

func (rn *RedisNode) ListRPush(key string, value []string) (int, error) {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_LRPUSH,
		key:       key,
		value:     value,
		replyChan: reply,
	}

	hValue, err := rn.waitResult(reply, key)
	if err == nil {
		if hValue == nil {
			return 0, fmt.Errorf("value is nil")
		} else {
			return hValue.(int), nil
		}
	}

	return 0, err
}

func (rn *RedisNode) ListRPushVariable(key string, value ...string) (int, error) {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_LRPUSH,
		key:       key,
		value:     value,
		replyChan: reply,
	}

	hValue, err := rn.waitResult(reply, key)
	if err == nil {
		if hValue == nil {
			return 0, fmt.Errorf("value is nil")
		} else {
			return hValue.(int), nil
		}
	}

	return 0, err
}

func (rn *RedisNode) ListLPushVariable(key string, value ...string) (int, error) {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_LLPUSH,
		key:       key,
		value:     value,
		replyChan: reply,
	}

	hValue, err := rn.waitResult(reply, key)
	if err == nil {
		if hValue == nil {
			return 0, fmt.Errorf("value is nil")
		} else {
			return hValue.(int), nil
		}
	}

	return 0, err
}

func (rn *RedisNode) ListRem(key string, v string, count int) (int, error) {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_LREM,
		key:       key,
		value:     strconv.Itoa(count),
		args:      []interface{}{v},
		replyChan: reply,
	}
	delCount, err := rn.waitResult(reply, key)
	if err == nil {
		return delCount.(int), nil
	}
	return 0, err
}

func (rn *RedisNode) ListTrim(key string, start, stop int) error {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_LTRIM,
		key:       key,
		args:      []interface{}{start, stop},
		replyChan: reply,
	}
	_, err := rn.waitResult(reply, key)
	return err
}

func (rn *RedisNode) ListLen(key string) (int, error) {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_LLEN,
		key:       key,
		replyChan: reply,
	}

	len, err := rn.waitResult(reply, key)
	if err == nil {
		return len.(int), nil
	}

	return 0, err
}

func (rn *RedisNode) ListLPop(key string) (string, error) {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_LPOP,
		key:       key,
		replyChan: reply,
	}

	value, err := rn.waitResult(reply, key)
	if err == nil {
		return value.(string), nil
	}

	return "", err
}

func (rn *RedisNode) HashSet(key string, hashV map[string]string) error {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_HSET,
		key:       key,
		value:     hashV,
		replyChan: reply,
	}
	_, err := rn.waitResult(reply, key)
	return err
}

func (rn *RedisNode) HashGet(key string) (map[string]string, error) {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_HGET,
		key:       key,
		replyChan: reply,
	}
	hValue, err := rn.waitResult(reply, key)
	if err == nil {
		return hValue.(map[string]string), nil
	}
	return nil, err
}

func (rn *RedisNode) HashGetFields(key string, fields ...string) ([]string, error) {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_HMGET,
		key:       key,
		value:     fields,
		replyChan: reply,
	}
	hValue, err := rn.waitResult(reply, key)
	if err == nil {
		if hValue == nil {
			return nil, nil
		} else {
			return hValue.([]string), nil
		}
	}

	return nil, err
}

func (rn *RedisNode) HashDel(key string, fields ...string) (int, error) {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_HDEL,
		key:       key,
		value:     fields,
		replyChan: reply,
	}
	delCount, err := rn.waitResult(reply, key)
	if err == nil {
		return delCount.(int), nil
	}
	return 0, err
}

func (rn *RedisNode) Exists(key string) (bool, error) {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_EXISTS,
		key:       key,
		replyChan: reply,
	}
	hValue, err := rn.waitResult(reply, key)
	if err == nil {
		return hValue.(bool), nil
	}
	return false, err
}

func NewRedisPipeLine() *RedisPipeLine {
	redisPipeLine := new(RedisPipeLine)
	redisPipeLine.paramSlice = make([]*RedisPipeLineParamS, 0)
	return redisPipeLine
}

func (rpl *RedisPipeLine) Append(containerType RedisContainerType, redisCmd string, key string, value ...interface{}) error {
	argsCount := len(value) + 1
	redisParam := &RedisPipeLineParamS{
		redisCmd:      redisCmd,
		containerType: containerType,
		args:          make([]interface{}, argsCount),
	}

	redisParam.args[0] = key
	var index = 1
	for index < argsCount {
		redisParam.args[index] = value[index-1]
		index++
	}
	rpl.paramSlice = append(rpl.paramSlice, redisParam)
	return nil
}

func (rpl *RedisPipeLine) Run(rn *RedisNode) (RedisPipeLineExecResult, error) {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_PIPELINE,
		key:       REDIS_PIPELINE.String(),
		value:     rpl.paramSlice,
		replyChan: reply,
	}
	res, err := rn.waitResult(reply, REDIS_PIPELINE.String())
	return res.(RedisPipeLineExecResult), err
}

func (rn *RedisNode) ClusterSlots() ([]*RedisClusterSlotS, error) {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_CLUSTERSLOTS,
		replyChan: reply,
	}
	hValue, err := rn.waitResult(reply, "ClusterSlots")
	if err == nil {
		return hValue.([]*RedisClusterSlotS), nil
	}
	return nil, err
}

func (rn *RedisNode) ScriptLoad(script []byte) (interface{}, error) {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_SCRIPT_LOAD,
		value:     script,
		replyChan: reply,
	}
	return rn.waitResult(reply, "script")
}

func (rn *RedisNode) Evalsha(args []interface{}) (interface{}, error) {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_EVAlSHA,
		args:      args,
		replyChan: reply,
	}
	return rn.waitResult(reply, "script")
}

func (rn *RedisNode) Incr(key string) (int, error) {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_INCR,
		key:       key,
		replyChan: reply,
	}

	hValue, err := rn.waitResult(reply, key)
	if err == nil {
		if hValue == nil {
			return 0, fmt.Errorf("value is nil")
		} else {
			return hValue.(int), nil
		}
	}

	return 0, err
}

func (rn *RedisNode) ZIncrBy(key, member string, increment int) (int, error) {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_ZINCRBY,
		key:       key,
		args:      []interface{}{strconv.Itoa(increment), member},
		replyChan: reply,
	}

	hValue, err := rn.waitResult(reply, key)
	if err == nil {
		if hValue == nil {
			return 0, fmt.Errorf("value is nil")
		} else {
			return hValue.(int), nil
		}
	}

	return 0, err
}

func (rn *RedisNode) ZScore(key, member string) (int, error) {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_ZSCORE,
		key:       key,
		args:      []interface{}{member},
		replyChan: reply,
	}

	hValue, err := rn.waitResult(reply, key)
	if err == nil {
		if hValue == nil {
			return 0, fmt.Errorf("member not exist")
		} else {
			return hValue.(int), nil
		}
	}

	return 0, err
}

func (rn *RedisNode) ZRem(key string, members ...string) (int, error) {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_ZREM,
		key:       key,
		value:     members,
		replyChan: reply,
	}

	hValue, err := rn.waitResult(reply, key)
	if err == nil {
		if hValue == nil {
			return 0, nil
		} else {
			return hValue.(int), nil
		}
	}

	return 0, err
}

func (rn *RedisNode) ZReverseRank(key string, member string) (int, error) {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_ZRRANK,
		key:       key,
		value:     member,
		replyChan: reply,
	}

	hValue, err := rn.waitResult(reply, key)
	if err == nil {
		if hValue == nil {
			return 0, nil
		} else {
			return hValue.(int), nil
		}
	}

	return 0, err
}

func (rn *RedisNode) Expire(key string, seconds int) error {
	reply := make(chan *RedisResultS)
	rn.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_EXPIRE,
		key:       key,
		value:     seconds,
		replyChan: reply,
	}

	_, err := rn.waitResult(reply, key)
	return err
}
