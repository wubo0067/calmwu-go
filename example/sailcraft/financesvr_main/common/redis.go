/*
 * @Author: calmwu
 * @Date: 2017-12-27 15:19:15
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-04-16 19:57:03
 * @Comment:
 */

package common

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sailcraft/base"
	"sailcraft/dataaccess/redistool"
	"sync"
)

var (
	GRedis        *redistool.RedisNode = nil
	redisInitOnce sync.Once
)

func InitRedis(redisSvrAddr string) error {
	var err error

	redisInitOnce.Do(func() {

		GRedis = redistool.NewRedis(redisSvrAddr, 5)
		err := GRedis.Start()
		if err != nil {
			base.GLog.Error("get RedisNode[%s] failed! reason[%s]", redisSvrAddr, err.Error())
		}

		base.GLog.Debug("Start RedisNode[%s] successed!", redisSvrAddr)
	})
	return err
}

func GetStrDataFromRedis(key string, objPtr interface{}) error {
	v := reflect.ValueOf(objPtr)
	if v.Kind() == reflect.Ptr {
		val, err := GRedis.StringGet(key)
		if err != nil {
			base.GLog.Error("Get Key[%s] data from redis failed! reason[%s]", key, err.Error())
			return err
		}

		if val == nil {
			err := fmt.Errorf("Get Key[%s] is not exist!", key)
			base.GLog.Error(err.Error())
			return err
		}

		if redisData, ok := val.([]byte); ok {
			err = json.Unmarshal(redisData, objPtr)
			if err != nil {
				base.GLog.Error("Unmarshal key[%s] failed! reason[%s]", key, err.Error())
				return err
			}
			return nil
		} else {
			err := fmt.Errorf("val type is not []byte")
			base.GLog.Error(err.Error())
			return err
		}
	}
	err := fmt.Errorf("objPtr Kind[%s] is not pointer", v.Kind().String())
	base.GLog.Error(err)
	return err
}
