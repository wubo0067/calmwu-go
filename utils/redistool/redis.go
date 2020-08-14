/*
 * @Author: calmwu
 * @Date: 2017-10-23 14:59:25
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-30 19:48:37
 * @Comment:
 */

package redistool

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis"
	utils "github.com/wubo0067/calmwu-go/utils"
)

type RedisMgr struct {
	redisSvrAddrs      []string
	isCluster          bool
	sessionCount       int
	password           string
	redisClient        *redis.Client
	redisClusterClient *redis.ClusterClient
	redisCmdable       redis.Cmdable
	redisCmdChan       chan *RedisCommandData
	redisExitChan      chan struct{}
	redisRoutineWG     sync.WaitGroup
}

// var (
// 	once sync.Once
// )

func NewRedisMgr(redisSvrAddrs []string, sessionCount int, isCluster bool, password string) *RedisMgr {
	redisMgr := new(RedisMgr)
	redisMgr.redisSvrAddrs = redisSvrAddrs
	redisMgr.sessionCount = sessionCount
	redisMgr.isCluster = isCluster
	redisMgr.password = password
	redisMgr.redisCmdChan = make(chan *RedisCommandData, 10240)
	redisMgr.redisExitChan = make(chan struct{})
	return redisMgr
}

func (rm *RedisMgr) GetClusterClient() *redis.ClusterClient {
	return rm.redisClusterClient
}

func (rm *RedisMgr) GetClient() *redis.Client {
	return rm.redisClient
}

func (rm *RedisMgr) Start() error {
	utils.ZLog.Debugf("redisSvrAddrs:%v", rm.redisSvrAddrs)

	if rm.isCluster {
		rm.redisClusterClient = redis.NewClusterClient(
			&redis.ClusterOptions{
				IdleTimeout: -1,
				Addrs:       rm.redisSvrAddrs,
				OnConnect: func(ctx context.Context, cn *redis.Conn) error {
					utils.ZLog.Infof("redis.Conn:%s", cn.String())
					return nil
				},
				Password:     rm.password,
				ReadOnly:     true,
				PoolSize:     rm.sessionCount,
				MinIdleConns: rm.sessionCount,
			},
		)
		rm.redisCmdable = rm.redisClusterClient
	} else {
		rm.redisClient = redis.NewClient(&redis.Options{
			Addr:     rm.redisSvrAddrs[0],
			Password: rm.password,
			OnConnect: func(ctx context.Context, conn *redis.Conn) error {
				utils.ZLog.Infof("redis.Conn:%s", conn.String())
				return nil
			},
			PoolSize:     rm.sessionCount,
			MinIdleConns: rm.sessionCount,
			IdleTimeout:  -1,
		})
		rm.redisCmdable = rm.redisClient
	}

	var index = 0
	for index < rm.sessionCount {
		// 启动命令处理协程
		go redisCmdPerformRoutine(rm)
		rm.redisRoutineWG.Add(1)
		index++
	}

	var err error
	var res string
	if rm.isCluster {
		res, err = rm.redisClusterClient.Ping(context.TODO()).Result()
		utils.ZLog.Debugf("redisClusterClient Ping:%s", res)
	} else {
		res, err = rm.redisClient.Ping(context.TODO()).Result()
		utils.ZLog.Debugf("redisClient Ping:%s", res)
	}

	return err
}

func (rm *RedisMgr) Stop() {
	close(rm.redisExitChan)
	rm.redisRoutineWG.Wait()
	if rm.isCluster {
		rm.redisClusterClient.Close()
	} else {
		rm.redisClient.Close()
	}
}

func redisCmdPerformRoutine(rm *RedisMgr) {
	utils.ZLog.Debug("redisCmdPerformRoutine running")

	defer func() {
		rm.redisRoutineWG.Done()
		if err := recover(); err != nil {
			// 回收panic，防止整个服务panic
			stackInfo := utils.CallStack(1)
			utils.ZLog.Warnw("redisCmdPerformRoutine panic recovered! err", err, "stack", string(stackInfo))
		}
	}()
L:
	for {
		select {
		case redisCmdData, ok := <-rm.redisCmdChan:

			if ok {
				switch redisCmdData.cmd {
				case REDIS_GET:
					val, err := rm.redisCmdable.Get(context.TODO(), redisCmdData.key).Result()
					if err != nil {
						utils.ZLog.Errorf("Get key[%s] failed! error:%s", redisCmdData.key, err.Error())
						redisCmdData.replyChan <- &RedisResultS{Ok: false, Result: err}
					} else {
						redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: val}
					}
				case REDIS_SET:
					// 设置key value
					expiration := time.Duration(redisCmdData.ttl) * time.Second
					_, err := rm.redisCmdable.Set(context.TODO(), redisCmdData.key, redisCmdData.value, expiration).Result()
					if err == nil {
						// set成功
						redisCmdData.replyChan <- &RedisResultS{Ok: true}
					} else {
						utils.ZLog.Errorf("Set key[%s] failed! error:%s", redisCmdData.key, err.Error())
						redisCmdData.replyChan <- &RedisResultS{Ok: false, Result: err}
					}
				case REDIS_DEL:
					res, err := rm.redisCmdable.Del(context.TODO(), redisCmdData.key).Result()
					utils.ZLog.Debugf("Del key[%s] res[%d]", redisCmdData.key, res)
					if err != nil {
						redisCmdData.replyChan <- &RedisResultS{Ok: false, Result: err}
					} else {
						redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: res}
					}
				case REDIS_SETNX:
					// 设置key value
					expiration := time.Duration(redisCmdData.ttl) * time.Second
					_, err := rm.redisCmdable.SetNX(context.TODO(), redisCmdData.key, redisCmdData.value, expiration).Result()
					if err == nil {
						// set成功
						redisCmdData.replyChan <- &RedisResultS{Ok: true}
					} else {
						utils.ZLog.Errorf("SetNX key[%s] failed! error:%s", redisCmdData.key, err.Error())
						redisCmdData.replyChan <- &RedisResultS{Ok: false, Result: err}
					}
				case REDIS_HGETALL:
					val, err := rm.redisCmdable.HGetAll(context.TODO(), redisCmdData.key).Result()
					if err == nil {
						// set成功
						redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: val}
					} else {
						utils.ZLog.Errorf("HGETALL key[%s] failed! error:%s", redisCmdData.key, err.Error())
						redisCmdData.replyChan <- &RedisResultS{Ok: false, Result: err}
					}
				case REDIS_HMGET:
				case REDIS_HMSET:
					res, err := rm.redisCmdable.HMSet(context.TODO(), redisCmdData.key, redisCmdData.value.(map[string]interface{})).Result()
					utils.ZLog.Debugf("HMSET key[%s] res:%v", redisCmdData.key, res)
					if err == nil {
						// set成功
						redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: res}
					} else {
						utils.ZLog.Errorf("HMSET key[%s] failed! error:%s", redisCmdData.key, err.Error())
						redisCmdData.replyChan <- &RedisResultS{Ok: false, Result: err}
					}
				case REDIS_HSET:
					res, err := rm.redisCmdable.HSet(context.TODO(), redisCmdData.key, redisCmdData.fieldKey, redisCmdData.value).Result()
					utils.ZLog.Debugf("HSET key[%s] fieldkey[%s] res:%v", redisCmdData.key, redisCmdData.fieldKey, res)
					if err == nil {
						// set成功
						redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: res}
					} else {
						utils.ZLog.Errorf("HSET key[%s] fieldKey[%s] failed! error:%s", redisCmdData.key,
							redisCmdData.fieldKey, err.Error())
						redisCmdData.replyChan <- &RedisResultS{Ok: false, Result: err}
					}
				case REDIS_HDEL:
				case REDIS_PIPELINE:
					cmds, err := rm.redisCmdable.Pipelined(context.TODO(), redisCmdData.value.(redisPipelineFn))
					if err != nil {
						utils.ZLog.Errorf("Pipelined failed! reason:%s", err.Error())
					}
					// if cmds != nil && len(cmds) > 0 {
					// 	utils.ZLog.Debugf("Pipelined cmds:%v", cmds)
					// }
					// 如果部分成功pipelined会返回err，这里外部去检查每个具体的cmd结果
					redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: cmds}
				case REDIS_EXISTS:
					val, err := rm.redisCmdable.Exists(context.TODO(), redisCmdData.key).Result()
					if err != nil {
						redisCmdData.replyChan <- &RedisResultS{Ok: false, Result: err}
					} else {
						redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: val}
					}
				case REDIS_ZINCRBY:
					// 当 key 不存在，或 member 不是 key 的成员时，相当于zadd
					val, err := rm.redisCmdable.ZIncrBy(context.TODO(), redisCmdData.key, redisCmdData.value.(float64), redisCmdData.fieldKey).Result()
					utils.ZLog.Debugf("ZIncrBy key[%s] increment[%v] memberKey[%s] val:%v", redisCmdData.key, redisCmdData.value,
						redisCmdData.fieldKey, val)
					if err != nil {
						redisCmdData.replyChan <- &RedisResultS{Ok: false, Result: err}
					} else {
						redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: val}
					}
				case REDIS_EXPIRE:
					res, err := rm.redisCmdable.Expire(context.TODO(), redisCmdData.key, redisCmdData.value.(time.Duration)).Result()
					utils.ZLog.Debugf("Expire key[%s] expiration:%v", redisCmdData.key, redisCmdData.value)
					if err != nil {
						redisCmdData.replyChan <- &RedisResultS{Ok: false, Result: err}
					} else {
						redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: res}
					}
				case REDIS_DBSIZE:
					res, err := rm.redisCmdable.DBSize(context.TODO()).Result()
					utils.ZLog.Debugf("DBSize count:%d", res)
					if err != nil {
						redisCmdData.replyChan <- &RedisResultS{Ok: false, Result: err}
					} else {
						redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: res}
					}
				case REDIS_SCAN:
					scanRes := new(RedisScanResult)
					var err error
					scanRes.keys, scanRes.cursor, err = rm.redisCmdable.Scan(context.TODO(), redisCmdData.args[0].(uint64),
						redisCmdData.args[1].(string),
						redisCmdData.args[2].(int64)).Result()
					utils.ZLog.Debugf("Scan keycount[%d] cursor[%d]", len(scanRes.keys), scanRes.cursor)
					if err != nil {
						redisCmdData.replyChan <- &RedisResultS{Ok: false, Result: err}
					} else {
						redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: scanRes}
					}
				case REDIS_CLUSTERSCAN:
					scanRes := new(RedisScanResult)
					var err error
					cmdable := redisCmdData.args[0].(redis.Cmdable)
					scanRes.keys, scanRes.cursor, err = cmdable.Scan(context.TODO(),
						redisCmdData.args[1].(uint64),
						redisCmdData.args[2].(string),
						redisCmdData.args[3].(int64)).Result()
					utils.ZLog.Debugf("Scan keycount[%d] cursor[%d] err:%v", len(scanRes.keys), scanRes.cursor, err)
					if err != nil {
						redisCmdData.replyChan <- &RedisResultS{Ok: false, Result: err}
					} else {
						redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: scanRes}
					}
				case REDIS_SADD:
					res, err := rm.redisCmdable.SAdd(context.TODO(), redisCmdData.key, redisCmdData.args...).Result()
					utils.ZLog.Debugf("SADD key[%s] res:%d", redisCmdData.key, res)
					if err != nil {
						redisCmdData.replyChan <- &RedisResultS{Ok: false, Result: err}
					} else {
						redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: res}
					}
				case REDIS_SREM:
					res, err := rm.redisCmdable.SRem(context.TODO(), redisCmdData.key, redisCmdData.args...).Result()
					utils.ZLog.Debugf("SREM key[%s] res:%d", redisCmdData.key, res)
					if err != nil {
						redisCmdData.replyChan <- &RedisResultS{Ok: false, Result: err}
					} else {
						redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: res}
					}
				case REDIS_SISMEMBER:
					res, err := rm.redisCmdable.SIsMember(context.TODO(), redisCmdData.key, redisCmdData.value).Result()
					utils.ZLog.Debugf("SISMEMBER key[%s] res:%v", redisCmdData.key, res)
					if err != nil {
						redisCmdData.replyChan <- &RedisResultS{Ok: false, Result: err}
					} else {
						redisCmdData.replyChan <- &RedisResultS{Ok: true, Result: res}
					}
				default:
					err := fmt.Errorf("Cmd[%s] does not support", redisCmdData.cmd.String())
					utils.ZLog.Errorf(err.Error())
					redisCmdData.replyChan <- &RedisResultS{Ok: false, Result: err}
				}
			}
		case <-rm.redisExitChan:
			break L
		}
	}
	utils.ZLog.Warnf("redisCmdPerformRoutine Exit")
	return
}

func (rm *RedisMgr) waitResult(reply <-chan *RedisResultS, tag string) (interface{}, error) {
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
		return nil, ErrTimeOut(tag)
	}
	return nil, ErrTimeOut(tag)
}

func (rm *RedisMgr) StringGet(key string) (interface{}, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_GET,
		key:       key,
		replyChan: reply,
	}
	return rm.waitResult(reply, key)
}

func (rm *RedisMgr) StringSet(key string, value []byte) error {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_SET,
		key:       key,
		value:     value,
		replyChan: reply,
	}
	_, err := rm.waitResult(reply, key)
	return err
}

// 0: key不存在，> 0 删除成功
func (rm *RedisMgr) DelKey(key string) (int64, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_DEL,
		key:       key,
		replyChan: reply,
	}

	hValue, err := rm.waitResult(reply, key)
	if err == nil {
		return hValue.(int64), nil
	}

	return 0, err
}

func (rm *RedisMgr) StringSetNX(key string, value []byte, ttl int) error {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_SETNX,
		key:       key,
		value:     value,
		ttl:       ttl,
		replyChan: reply,
	}
	_, err := rm.waitResult(reply, key)
	return err
}

func (rm *RedisMgr) ListSet(key string, value []string) error {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_LSET,
		key:       key,
		value:     value,
		replyChan: reply,
	}
	_, err := rm.waitResult(reply, key)
	return err
}

func (rm *RedisMgr) ListGet(key string) ([]string, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_LGET,
		key:       key,
		replyChan: reply,
	}
	lValue, err := rm.waitResult(reply, key)
	if err == nil {
		return lValue.([]string), nil
	}
	return nil, err
}

func (rm *RedisMgr) ListRPush(key string, value []string) (int, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_LRPUSH,
		key:       key,
		value:     value,
		replyChan: reply,
	}

	hValue, err := rm.waitResult(reply, key)
	if err == nil {
		if hValue == nil {
			return 0, fmt.Errorf("value is nil")
		} else {
			return hValue.(int), nil
		}
	}

	return 0, err
}

func (rm *RedisMgr) ListRPushVariable(key string, value ...string) (int, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_LRPUSH,
		key:       key,
		value:     value,
		replyChan: reply,
	}

	hValue, err := rm.waitResult(reply, key)
	if err == nil {
		if hValue == nil {
			return 0, fmt.Errorf("value is nil")
		} else {
			return hValue.(int), nil
		}
	}

	return 0, err
}

func (rm *RedisMgr) ListLPushVariable(key string, value ...string) (int, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_LLPUSH,
		key:       key,
		value:     value,
		replyChan: reply,
	}

	hValue, err := rm.waitResult(reply, key)
	if err == nil {
		if hValue == nil {
			return 0, fmt.Errorf("value is nil")
		} else {
			return hValue.(int), nil
		}
	}

	return 0, err
}

func (rm *RedisMgr) ListRem(key string, v string, count int) (int, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_LREM,
		key:       key,
		value:     strconv.Itoa(count),
		args:      []interface{}{v},
		replyChan: reply,
	}
	delCount, err := rm.waitResult(reply, key)
	if err == nil {
		return delCount.(int), nil
	}
	return 0, err
}

func (rm *RedisMgr) ListTrim(key string, start, stop int) error {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_LTRIM,
		key:       key,
		args:      []interface{}{start, stop},
		replyChan: reply,
	}
	_, err := rm.waitResult(reply, key)
	return err
}

func (rm *RedisMgr) ListLen(key string) (int, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_LLEN,
		key:       key,
		replyChan: reply,
	}

	len, err := rm.waitResult(reply, key)
	if err == nil {
		return len.(int), nil
	}

	return 0, err
}

func (rm *RedisMgr) ListLPop(key string) (string, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_LPOP,
		key:       key,
		replyChan: reply,
	}

	value, err := rm.waitResult(reply, key)
	if err == nil {
		return value.(string), nil
	}

	return "", err
}

func (rm *RedisMgr) HashMSet(key string, hashV map[string]interface{}) error {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_HMSET,
		key:       key,
		value:     hashV,
		replyChan: reply,
	}
	_, err := rm.waitResult(reply, key)
	return err
}

func (rm *RedisMgr) HashSet(key string, fieldKey string, value interface{}) error {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_HSET,
		key:       key,
		fieldKey:  fieldKey,
		value:     value,
		replyChan: reply,
	}
	_, err := rm.waitResult(reply, key)
	return err
}

func (rm *RedisMgr) HashGetAll(key string) (map[string]string, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_HGETALL,
		key:       key,
		replyChan: reply,
	}
	hValue, err := rm.waitResult(reply, key)
	if err == nil {
		return hValue.(map[string]string), nil
	}
	return nil, err
}

func (rm *RedisMgr) HashDel(key string, fields ...string) (int, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_HDEL,
		key:       key,
		value:     fields,
		replyChan: reply,
	}
	delCount, err := rm.waitResult(reply, key)
	if err == nil {
		return delCount.(int), nil
	}
	return 0, err
}

// 0: 不存在 1: 存在
func (rm *RedisMgr) Exists(key string) (int64, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_EXISTS,
		key:       key,
		replyChan: reply,
	}
	hValue, err := rm.waitResult(reply, key)
	if err == nil {
		return hValue.(int64), nil
	}
	return 0, err
}

func (rm *RedisMgr) ScriptLoad(script []byte) (interface{}, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_SCRIPT_LOAD,
		value:     script,
		replyChan: reply,
	}
	return rm.waitResult(reply, "script")
}

func (rm *RedisMgr) Evalsha(args []interface{}) (interface{}, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_EVAlSHA,
		args:      args,
		replyChan: reply,
	}
	return rm.waitResult(reply, "script")
}

func (rm *RedisMgr) Incr(key string) (int, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_INCR,
		key:       key,
		replyChan: reply,
	}

	hValue, err := rm.waitResult(reply, key)
	if err == nil {
		if hValue == nil {
			return 0, fmt.Errorf("value is nil")
		} else {
			return hValue.(int), nil
		}
	}

	return 0, err
}

func (rm *RedisMgr) ZIncrBy(key, member string, increment int) (int, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_ZINCRBY,
		key:       key,
		args:      []interface{}{strconv.Itoa(increment), member},
		replyChan: reply,
	}

	hValue, err := rm.waitResult(reply, key)
	if err == nil {
		if hValue == nil {
			return 0, fmt.Errorf("value is nil")
		} else {
			return hValue.(int), nil
		}
	}

	return 0, err
}

func (rm *RedisMgr) ZScore(key, member string) (int, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_ZSCORE,
		key:       key,
		args:      []interface{}{member},
		replyChan: reply,
	}

	hValue, err := rm.waitResult(reply, key)
	if err == nil {
		if hValue == nil {
			return 0, fmt.Errorf("member not exist")
		} else {
			return hValue.(int), nil
		}
	}

	return 0, err
}

func (rm *RedisMgr) ZRem(key string, members ...string) (int, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_ZREM,
		key:       key,
		value:     members,
		replyChan: reply,
	}

	hValue, err := rm.waitResult(reply, key)
	if err == nil {
		if hValue == nil {
			return 0, nil
		} else {
			return hValue.(int), nil
		}
	}

	return 0, err
}

func (rm *RedisMgr) ZReverseRank(key string, member string) (int, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_ZRRANK,
		key:       key,
		value:     member,
		replyChan: reply,
	}

	hValue, err := rm.waitResult(reply, key)
	if err == nil {
		if hValue == nil {
			return 0, nil
		} else {
			return hValue.(int), nil
		}
	}

	return 0, err
}

func (rm *RedisMgr) Expire(key string, seconds int) (bool, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_EXPIRE,
		key:       key,
		value:     time.Duration(seconds) * time.Second,
		replyChan: reply,
	}

	ret, err := rm.waitResult(reply, key)
	return ret.(bool), err
}

func (rm *RedisMgr) Pipelined(fn redisPipelineFn) ([]redis.Cmder, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_PIPELINE,
		key:       "Pipelined",
		value:     fn,
		replyChan: reply,
	}

	hValue, err := rm.waitResult(reply, "Pipelined")
	if err == nil {
		return hValue.([]redis.Cmder), nil
	}

	return nil, err
}

func (rm *RedisMgr) DBSize() (int64, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_DBSIZE,
		replyChan: reply,
	}

	hValue, err := rm.waitResult(reply, "DBSize")
	if err == nil {
		return hValue.(int64), nil
	}

	return 0, err
}

func (rm *RedisMgr) Scan(cursor uint64, match string, count int64) ([]string, uint64, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_SCAN,
		args:      []interface{}{cursor, match, count},
		replyChan: reply,
	}

	hValue, err := rm.waitResult(reply, "Scan")
	if err == nil {
		scanRes := hValue.(*RedisScanResult)
		return scanRes.keys, scanRes.cursor, nil
	}

	return nil, 0, err
}

func (rm *RedisMgr) ClusterScan(masterNodeCmdable redis.Cmdable, cursor uint64, match string, count int64) ([]string, uint64, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_CLUSTERSCAN,
		args:      []interface{}{masterNodeCmdable, cursor, match, count},
		replyChan: reply,
	}

	hValue, err := rm.waitResult(reply, "Scan")
	if err == nil {
		scanRes := hValue.(*RedisScanResult)
		return scanRes.keys, scanRes.cursor, nil
	}

	return nil, 0, err
}

func (rm *RedisMgr) SAdd(key string, members ...interface{}) (int64, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_SADD,
		key:       key,
		args:      members,
		replyChan: reply,
	}

	hValue, err := rm.waitResult(reply, key)
	if err == nil {
		if hValue == nil {
			return 0, nil
		} else {
			return hValue.(int64), nil
		}
	}

	return 0, err
}

func (rm *RedisMgr) SRem(key string, members ...interface{}) (int64, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_SREM,
		key:       key,
		args:      members,
		replyChan: reply,
	}

	hValue, err := rm.waitResult(reply, key)
	if err == nil {
		if hValue == nil {
			return 0, nil
		} else {
			return hValue.(int64), nil
		}
	}
	return 0, err
}

func (rm *RedisMgr) SIsMember(key string, member interface{}) (bool, error) {
	reply := make(chan *RedisResultS)
	rm.redisCmdChan <- &RedisCommandData{
		cmd:       REDIS_SISMEMBER,
		key:       key,
		value:     member,
		replyChan: reply,
	}

	hValue, err := rm.waitResult(reply, key)
	if err == nil {
		return hValue.(bool), nil
	}

	return false, err
}
