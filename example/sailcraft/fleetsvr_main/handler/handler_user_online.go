package handler

import (
	"fmt"
	"sailcraft/fleetsvr_main/model"
)

const (
	ONLINE_KEY_UIN_TO_PID = "uin2pid."
)

func getUserOnlineKey(uin int) string {
	return fmt.Sprintf("%s%d", ONLINE_KEY_UIN_TO_PID, uin)
}

func IsUserOnline(uin int) (result bool, err error) {
	result = false
	err = nil

	if uin > 0 {
		key := getUserOnlineKey(uin)
		redis := model.GetSingletonRedis()
		if redis != nil {
			value, err := redis.Exists(key)
			if err == nil {
				result = value
			}
		}
	}

	return result, err
}
