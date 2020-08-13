/*
 * @Author: calmwu
 * @Date: 2017-10-26 15:38:10
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-21 11:27:20
 * @Comment:
 */
package redistool

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/wubo0067/calmwu-go/utils"
	"go.uber.org/zap/zapcore"
)

const (
	redisSvrAddrs = "192.168.68.228:7000,192.168.68.228:7001,192.168.68.229:7002,192.168.68.229:7003,192.168.68.230:7004,192.168.68.230:7005"
)

func createRedisMgr(t *testing.T) (*RedisMgr, error) {
	redisMgr := NewRedisMgr(strings.Split(redisSvrAddrs, ","), 10, true, "")
	err := redisMgr.Start()
	if err != nil {
		t.Errorf("redisMgr[%s] start failed! reason:%s", redisSvrAddrs,
			err.Error())
		return nil, err
	}
	t.Log("Create NewRedisMgr successed!")
	return redisMgr, nil
}

//go test -v -run TestSetValue
func TestSetValue(t *testing.T) {
	utils.InitDefaultZapLog("test.log", zapcore.DebugLevel)

	redisMgr, err := createRedisMgr(t)
	if err != nil {
		t.Error(err.Error())
		return
	}

	start := time.Now()
	for i := 1000; i < 2000; i++ {
		key := fmt.Sprintf("doyo%08d", i)
		val := fmt.Sprintf("val-%s", key)
		err = redisMgr.StringSet(key, []byte(val))
		if err != nil {
			t.Errorf("StringSet %s failed! reason:%s", key, err.Error())
		}
	}
	elapsed := time.Since(start)
	t.Logf("elapsed:%s", elapsed)

	dbSize, err := redisMgr.DBSize()
	t.Logf("dbSize:%d", dbSize)

	redisMgr.Stop()
}

func TestScan(t *testing.T) {
	utils.InitDefaultZapLog("test.log", zapcore.DebugLevel)

	redisMgr, err := createRedisMgr(t)
	if err != nil {
		t.Error(err.Error())
		return
	}

	dbSize, err := redisMgr.DBSize()
	t.Logf("dbSize:%d", dbSize)

	var batchCount int64 = 100
	var cursor uint64 = 0

	// cluster scan
	redisMgr.GetClusterClient().ForEachMaster(func(master *redis.Client) error {
		for {
			keys, newCursor, err := redisMgr.ClusterScan(master, cursor, "", batchCount)
			if err != nil {
				t.Error(err.Error())
				return nil
			}

			utils.ZLog.Debugf("---newCursor:%d keyCount:%d", newCursor, len(keys))

			for i := range keys {
				utils.ZLog.Debug(keys[i])
			}

			if newCursor == 0 {
				utils.ZLog.Debug("Scan completed! newCursor=0")
				return nil
			}
			cursor = newCursor
		}
		return nil
	})

}

func TestPipe(t *testing.T) {
	utils.InitDefaultZapLog("test.log", zapcore.DebugLevel)

	redisMgr, err := createRedisMgr(t)
	if err != nil {
		t.Error(err.Error())
		return
	}

	cmds, err := redisMgr.Pipelined(func(pipe redis.Pipeliner) error {
		val, err := pipe.Get("doyo00001551").Result()
		if err != nil {
			utils.ZLog.Debugf("doyo00001551 err:%s", err.Error())
		} else {
			utils.ZLog.Debugf("doyo00001551 val:%s", val)
		}

		pipe.Get("doyo00001552")
		pipe.Get("doyo00001553")
		pipe.Get("doyo0000")
		return nil
	})

	// utils.ZLog.Debugf("err:%v", err)
	utils.ZLog.Debugf("cmds:%+v", cmds)

	for index := range cmds {
		cmd := cmds[index].(*redis.StringCmd)
		name := cmd.Name()
		key := cmd.Args()[1].(string)
		if cmd.Err() == nil {
			utils.ZLog.Debugf("%s %s %s", name, key, cmd.Val())
		} else {
			utils.ZLog.Debugf("%s %s failed! reason:%s", name, key, cmd.Err().Error())
		}
	}
}
