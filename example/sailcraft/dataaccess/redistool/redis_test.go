/*
 * @Author: calmwu
 * @Date: 2017-10-26 15:38:10
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-10-27 17:26:21
 * @Comment:
 */
package redistool

import (
	"fmt"
	"testing"
	"time"
)

func createRedisMgr() (*RedisNode, error) {
	redisAddr := "123.59.40.19:7003"
	sessionCount := 5
	redisMgr := NewRedis(redisAddr, sessionCount)
	err := redisMgr.Start()
	return redisMgr, err
}

func TestSetGetValue(t *testing.T) {
	redisMgr, err := createRedisMgr()
	if err != nil {
		t.Error(err.Error())
		return
	}

	defer redisMgr.Stop()

	key := "AocDisplay"
	value := "Aoc 32 QHD HDMI MHL PIP PBP FREE"

	err = redisMgr.StringSet(key, []byte(value))
	if err != nil {
		t.Error(err.Error())
		return
	} else {
		t.Logf("set key[%s] value[%s] successed!\n", key, value)
	}

	result, err := redisMgr.StringGet(key)
	if err != nil {
		t.Error(err.Error())
		return
	}

	t.Log("get value:", string(result.([]byte)))
}

func TestListSetGet(t *testing.T) {
	redisMgr, err := createRedisMgr()
	if err != nil {
		t.Error(err.Error())
		return
	}

	defer redisMgr.Stop()

	lKey := "china"
	lValue := make([]string, 10)
	i := 0
	for i < 10 {
		lValue[i] = fmt.Sprintf("hello_%d", i)
		i++
	}

	fmt.Println(lValue)

	err = redisMgr.ListSet(lKey, lValue)
	if err != nil {
		t.Error(err.Error())
		return
	} else {
		t.Logf("lset key[%s] value[%v] successed!\n", lKey, lValue)
	}

	lValue, err = redisMgr.ListGet(lKey)
	if err != nil {
		t.Error(err.Error())
		return
	} else {
		t.Logf("lget key[%s] value[%v] successed!\n", lKey, lValue)
	}
}

func TestHashSetGet(t *testing.T) {
	type Server struct {
		Name    string  `xorm:"varchar(128) default('') not null 'name'"`
		ID      int32   `xorm:"int 'id'"`
		Enabled bool    `xorm:"bool default('') 'enable'"`
		Fnum1   float32 `xorm:"bool 'fnum1'"`
		Fnum2   float64 `xorm:"bool 'fnum2'"`
		Date    time.Time
	}

	server := &Server{
		Name:    "Arslan",
		ID:      123456,
		Enabled: true,
		Fnum1:   3.14,
		Fnum2:   3.1515126,
		Date:    time.Now(),
	}

	mapV, _ := ConvertObjToRedisHash(server)
	t.Log(mapV)

	redisMgr, err := createRedisMgr()
	if err != nil {
		t.Error(err.Error())
		return
	}
	defer redisMgr.Stop()

	hKey := "Server"
	err = redisMgr.HashSet(hKey, mapV)
	if err != nil {
		t.Error(err.Error())
		return
	} else {
		t.Logf("hset key[%s] value:%v successed!\n", hKey, mapV)
	}

	mapV, err = redisMgr.HashGet(hKey)
	if err != nil {
		t.Error(err.Error())
		return
	} else {
		t.Logf("hget key[%s] value:%v successed!\n", hKey, mapV)
	}

	hObj := new(Server)
	err = ConvertRedisHashToObj(mapV, hObj)
	if err != nil {
		t.Error(err.Error())
		return
	} else {
		t.Log(hObj)
	}
}

func TestConvertSliceToRedisList(t *testing.T) {
	strSlice := make([]string, 10)
	strSlice[0] = "1"
	strSlice[2] = "2"
	_, err := ConvertSliceToRedisList(strSlice)
	if err != nil {
		t.Error(err.Error())
	}

	numSlice := make([]int, 10)
	numSlice[0] = 1
	numSlice[4] = 2
	redisL, _ := ConvertSliceToRedisList(numSlice)
	t.Log(redisL)
}

func TestConvertRedisListToSlice(t *testing.T) {
	strSlice := make([]string, 10)
	strSlice[0] = "1"
	strSlice[2] = "2"

	slice := make([]int, len(strSlice))
	err := ConvertRedisListToSlice(strSlice, slice)
	if err != nil {
		t.Error(err.Error())
		return
	} else {
		t.Log(slice)
	}

	strSlice = make([]string, 10)
	strSlice[0] = "3.124"
	strSlice[2] = "9.8"

	fSlice := make([]float32, len(strSlice))
	err = ConvertRedisListToSlice(strSlice, fSlice)
	if err != nil {
		t.Error(err.Error())
		return
	} else {
		t.Log(fSlice)
	}
}

func TestPipeLine(t *testing.T) {
	redisMgr, err := createRedisMgr()
	if err != nil {
		t.Error(err.Error())
		return
	}
	defer redisMgr.Stop()

	redisPipeLine := NewRedisPipeLine()
	redisPipeLine.Append(REDIS_CONTAINER_STR, "GET", "cpu")
	redisPipeLine.Append(REDIS_CONTAINER_STR, "SET", "cpu", "ARM")

	redisPipeLine.Append(REDIS_CONTAINER_HASH, "HGETALL", "Server")
	redisPipeLine.Append(REDIS_CONTAINER_HASH, "HSET", "Server", "id", "++++++++++++++######################")
	redisPipeLine.Append(REDIS_CONTAINER_HASH, "HGETALL", "Server")

	redisPipeLine.Append(REDIS_CONTAINER_LIST, "RPUSH", "book", "c++")
	redisPipeLine.Append(REDIS_CONTAINER_LIST, "RPUSH", "book", "java")
	redisPipeLine.Append(REDIS_CONTAINER_LIST, "RPUSH", "book", "golang")
	redisPipeLine.Append(REDIS_CONTAINER_LIST, "LRANGE", "book", 0, -1)

	pipeLineRes, err := redisPipeLine.Run(redisMgr)
	if err != nil {
		t.Error(err.Error())
	} else {
		for index, _ := range pipeLineRes {
			res := pipeLineRes[index]

			t.Logf("cmd[%s] key[%s] container[%d] replyType[%d]", res.RedisCmd, res.Key, res.ContainerType,
				res.reply.Type)

			switch res.RedisCmd {
			case "GET":
				{
					str, _ := res.String()
					t.Log(str)
				}
			case "HGETALL":
				{
					h, _ := res.Hash()
					t.Log(h)
				}
			case "LRANGE":
				{
					l, _ := res.List()
					t.Log(l)
				}
			}
		}
	}
}

func TestGetClusterSlots(t *testing.T) {
	redisMgr, err := createRedisMgr()
	if err != nil {
		t.Error(err.Error())
		return
	}
	defer redisMgr.Stop()

	slots, err := redisMgr.ClusterSlots()
	if err != nil {
		t.Error(err.Error())
	} else {
		for i := range slots {
			t.Logf("%v", slots[i])
		}
	}
}

func createRedisNodes() []*RedisNode {
	redisNodes := make([]*RedisNode, 3)

	redisNodes[0] = NewRedis("123.59.40.19:7003", 5)
	redisNodes[0].Start()

	redisNodes[1] = NewRedis("123.59.40.19:7004", 5)
	redisNodes[1].Start()

	redisNodes[2] = NewRedis("123.59.40.19:7005", 5)
	redisNodes[2].Start()

	return redisNodes
}

func TestCluster(t *testing.T) {
	nodes := createRedisNodes()

	cluster, err := GetRedisCluster(nodes)
	if err != nil {
		t.Error(err.Error())
	} else {
		rKey := "Server"
		node, err := cluster.GetRedisNodeByKey(rKey)
		if err != nil {
			t.Error(err.Error())
		} else {
			t.Logf("%v", node)
		}

		rKey = "User.cache.101101"
		node, err = cluster.GetRedisNodeByKey(rKey)
		if err != nil {
			t.Error(err.Error())
		} else {
			t.Logf("%v", node)
		}
	}
}
