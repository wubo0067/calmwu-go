package model

import (
	"encoding/json"
	"fmt"
	"sailcraft/dataaccess/redistool"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/table"
)

const (
	TABLE_NAME_GUILD_LEAVE_INFO       = "guild.leave"
	GUILD_LEAVE_INFO_REDIS_CACHED_KEY = "cache"
)

type GuildLeaveInfoModel struct {
	Uin int
}

func (this *GuildLeaveInfoModel) CachedKey() string {
	return fmt.Sprintf("%s.%s.%d", TABLE_NAME_GUILD_LEAVE_INFO, GUILD_LEAVE_INFO_REDIS_CACHED_KEY, this.Uin)
}

func (this *GuildLeaveInfoModel) SetGuildLeaveInfo(leaveInfo *table.TblGuildLeaveInfo) (int, error) {
	if this.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	if leaveInfo == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null point")
	}

	data, err := json.Marshal(leaveInfo)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisMgr[%s] is nil", redisKey)
	}

	err = redisMgr.StringSet(redisKey, data)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *GuildLeaveInfoModel) GetGuildLeaveInfo() (*table.TblGuildLeaveInfo, error) {
	if this.Uin <= 0 {
		return nil, custom_errors.New("uin is invalid")
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return nil, custom_errors.New("redisMgr[%s] is nil", redisKey)
	}

	i, err := redisMgr.StringGet(redisKey)
	if err != nil && !redistool.IsKeyNotExist(err) {
		return nil, err
	}

	if i == nil {
		return nil, nil
	}

	if data, ok := i.([]byte); ok {
		leaveInfo := new(table.TblGuildLeaveInfo)
		err = json.Unmarshal(data, leaveInfo)
		if err == nil {
			return leaveInfo, nil
		}
		// 如果解析失败，视为没有退出信息
	}

	return nil, nil
}
