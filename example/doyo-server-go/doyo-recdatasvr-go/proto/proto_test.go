/*
 * @Author: calmwu
 * @Date: 2018-10-26 15:37:52
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-29 19:24:42
 */

package proto

import (
	"testing"
	"time"

	"github.com/pquerna/ffjson/ffjson"
)

func TestDoyoRedisUserInfo(t *testing.T) {
	var redisUser DoyoRedisUserInfo
	redisUser.UserID = "doyo12345"
	redisUser.UserCountry = "CN"
	redisUser.UserLanguage = "zh"
	redisUser.LogoutTime = time.Now()

	serialData, err := ffjson.Marshal(&redisUser)
	if err != nil {
		t.Error(err.Error())
		return
	}

	t.Log(string(serialData))

	var newRedisUser DoyoRedisUserInfo
	err = ffjson.Unmarshal(serialData, &newRedisUser)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestDoyoRecDataCmd(t *testing.T) {
	t.Logf("DoyoRecDataCmdUserLogin=%d", DoyoRecDataCmdUserLogin)
	t.Logf("DoyoRecDataCmdUserLogout=%d", DoyoRecDataCmdUserLogout)
}
