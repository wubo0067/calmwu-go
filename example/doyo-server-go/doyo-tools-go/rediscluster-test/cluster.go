/*
 * @Author: calmwu
 * @Date: 2018-10-26 15:54:15
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-10-31 17:40:27
 */

package main

import (
	base "doyo-server-go/doyo-base-go"
	"time"

	"github.com/go-redis/redis"
)

func main() {
	logger := base.NewSimpleLog(nil)

	redisdb := redis.NewClusterClient(&redis.ClusterOptions{
		// Addrs: []string{"192.168.68.228:7000", "192.168.68.228:7001", "192.168.68.229:7002",
		// 	"192.168.68.229:7003", "192.168.68.230:7004", "192.168.68.230:7005"},
		Addrs: []string{"192.168.68.228:7000"},
		OnConnect: func(conn *redis.Conn) error {
			// 会连接所有的master
			callstack := base.CallStack(1)
			logger.Printf("---------conn:%s, callstack:%s", conn.String(), callstack)
			return nil
		},
		ReadOnly: true, // 在从库上开启readonly
	})

	logger.Printf("Option:%+v\n", redisdb.Options())

	err := redisdb.ReloadState()
	if err != nil {
		logger.Printf("err:%+s\n", err.Error())
		return
	}

	status := redisdb.Ping()
	logger.Println(status.String())

	res, err := redisdb.ClusterInfo().Result()
	if err != nil {
		logger.Println(err.Error())
	} else {
		logger.Printf("ClusterInfo res:%s\n", res)
	}

	redisdb.ForEachMaster(func(master *redis.Client) error {
		logger.Printf("master %s", master.String())
		return nil
	})

	err = redisdb.ForEachSlave(func(slave *redis.Client) error {
		logger.Printf("slave %s", slave.String())
		return nil
	})

	logger.Printf("%+v", redisdb.PoolStats())

	exist, err := redisdb.Exists("calmwu").Result()
	if err != nil {
		logger.Printf("err:%s\n", err.Error())
	} else {
		logger.Printf("calmwu res:%d\n", exist)
	}

	res, err = redisdb.Set("calmwu", "ryzen", 0).Result()
	if err != nil {
		logger.Printf("err:%s\n", err.Error())
	} else {
		logger.Printf("res:%s\n", res)
	}

	var cmdable redis.Cmdable = redisdb

	if _, ok := cmdable.(*redis.Client); !ok {
		logger.Printf("client is not redis.Client")
	}

	if _, ok := cmdable.(*redis.ClusterClient); ok {
		logger.Printf("client is redis.ClusterClient")
	}

	res, err = cmdable.Get("calmwu").Result()
	if err != nil {
		logger.Printf("err:%s\n", err.Error())
	} else {
		logger.Printf("res:%s\n", res)
	}

	res, err = cmdable.Get("noexist").Result()
	if err == redis.Nil {
		logger.Printf("err:%s\n", err.Error())
	} else {
		logger.Printf("res:%s\n", res)
	}

	// 使用pipeline来降低redis的操作
	cmds, err := cmdable.Pipelined(func(pipe redis.Pipeliner) error {
		// 查询key是否存在
		val, err := pipe.Get("xiaomi").Result()
		if err != nil {
			logger.Printf("err:%s\n", err.Error())
			pipe.Set("xiaomi", "leijun", 0)
		} else {
			logger.Printf("key:xiaomi value:%s\n", val)
			pipe.Incr("xiaomi")
			pipe.Incr("xiaomi")
			pipe.Incr("xiaomi")
		}
		return nil
	})

	logger.Printf("cmds:%v\n", cmds)

	time.Sleep(30 * time.Second)

	redisdb.Close()
}
