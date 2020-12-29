package model

import (
	"sailcraft/dataaccess/redistool"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/table"
)

const (
	TABLE_NAME_GUILD_WAR = "guildwar"
)

type GuildWarModel struct {
}

func (this *GuildWarModel) CachedKey() string {
	return TABLE_NAME_GUILD_WAR
}

func (this *GuildWarModel) Query() (*table.TblGuildWar, error) {
	redisKey := this.CachedKey()
	redisMgr := GetSingletonRedis()
	if redisMgr == nil {
		return nil, custom_errors.New("redisMgr is nil")
	}

	mapKV, err := redisMgr.HashGet(redisKey)
	if err != nil {
		return nil, err
	}

	record := new(table.TblGuildWar)
	err = redistool.ConvertRedisHashToObj(mapKV, record)
	if err != nil {
		return nil, err
	}

	return record, nil
}
