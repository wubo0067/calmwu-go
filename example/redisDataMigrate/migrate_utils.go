package main

import (
	"fmt"
	l4g "log4go"
	"sync"

	"time"

	"strings"

	"github.com/fzzy/radix/extra/pool"
	"github.com/fzzy/radix/redis"
)

type RedisConnPoolMgr struct {
	RedisConnPoolMap map[string]*pool.Pool
	Monitor          *sync.RWMutex
}

var (
	gLog              l4g.Logger        = make(l4g.Logger)
	gRedisConnPoolMgr *RedisConnPoolMgr = nil
)

func InitLog(logname string) {

	logfile := fmt.Sprintf("%s", logname)
	log_writer := l4g.NewFileLogWriter(logfile, false)
	log_writer.SetRotate(true)
	log_writer.SetRotateSize(50 * 1024 * 1024)
	log_writer.SetRotateMaxBackup(10)
	gLog.AddFilter("normal", l4g.FINE, log_writer)
	return
}

func (poolMgr *RedisConnPoolMgr) AddRedisPool(addr string, redisSvrAuth string) error {
	_, ok := poolMgr.RedisConnPoolMap[addr]
	if !ok {
		df := func(network, addr string) (*redis.Client, error) {
			client, err := redis.Dial(network, addr)
			if err != nil {
				gLog.Error("Connect to Redis[%s] failed! error[%s]", addr, err.Error())
				return nil, err
			}
			// 判断是否需要密码校验
			gLog.Debug("Auth Redis[%s] [%s]", addr, redisSvrAuth)
			if !strings.EqualFold(redisSvrAuth, "nil") && len(redisSvrAuth) > 0 {
				if err = client.Cmd("AUTH", redisSvrAuth).Err; err != nil {
					gLog.Debug("Auth Redis[%s] failed! error[%s]", addr, err.Error())
					client.Close()
					return nil, err
				} else {
					gLog.Debug("Auth Redis[%s] successed!", addr)
				}
			}
			return client, nil
		}
		// 不存在
		RedisConnPool, err := pool.NewCustomPool("tcp", addr, 5, df)
		if err != nil {
			gLog.Error("Create Redis Pool[%s] failed! err[%s]", addr, err.Error())
			return err
		} else {
			poolMgr.RedisConnPoolMap[addr] = RedisConnPool
			gLog.Debug("Create Redis Pool[%s] successed!", addr)
		}

	} else {
		gLog.Error("Redis Connect Pool[%s] is already exists!", addr)
	}
	return nil
}

func (poolMgr *RedisConnPoolMgr) GetRedisConn(addr string) *redis.Client {
	RedisConnPool, ok := poolMgr.RedisConnPoolMap[addr]
	if ok {
		conn, err := RedisConnPool.Get()
		if err != nil {
			gLog.Error("Get Redis[%s] connection from pool failed!", addr)
			return nil
		} else {
			//gLog.Debug("Get Redis[%s] connection from pool successed!", addr)
			return conn
		}
	}
	gLog.Error("Redis Connect Pool[%s] is not exist!", addr)
	return nil
}

func (poolMgr *RedisConnPoolMgr) PutRedisConn(addr string, conn *redis.Client) {
	if conn != nil {
		RedisConnPool, ok := poolMgr.RedisConnPoolMap[addr]
		if ok {
			RedisConnPool.Put(conn)
			//gLog.Debug("Put Redis[%s] connection to pool successed!", addr)
			return
		}
	}
	gLog.Error("Put Redis[%s] connection to pool failed!", addr)
}

func (poolMgr *RedisConnPoolMgr) Clean() {
	for addr, RedisConnPool := range poolMgr.RedisConnPoolMap {
		gLog.Debug("now close Redis connect pool[%s]", addr)
		RedisConnPool.Empty()
		time.Sleep(time.Second)
	}
}

func InitRedisConnPools(config *Config) error {
	if gRedisConnPoolMgr == nil {
		gRedisConnPoolMgr = new(RedisConnPoolMgr)
		gRedisConnPoolMgr.RedisConnPoolMap = make(map[string]*pool.Pool)
		gRedisConnPoolMgr.Monitor = new(sync.RWMutex)
	}

	// 为每个redis建立连接池
	for index, _ := range config.MigrateFromServs {
		err := gRedisConnPoolMgr.AddRedisPool(config.MigrateFromServs[index].RedisSvrAddr,
			config.MigrateFromServs[index].RedisSvrAuth)
		if err != nil {
			return err
		}
	}

	for index, _ := range config.MigrateToServs {
		err := gRedisConnPoolMgr.AddRedisPool(config.MigrateToServs[index].RedisSvrAddr,
			config.MigrateToServs[index].RedisSvrAuth)
		if err != nil {
			return err
		}
	}
	return nil
}
