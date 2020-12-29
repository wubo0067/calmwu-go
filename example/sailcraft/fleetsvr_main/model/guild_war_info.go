package model

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/dataaccess/redistool"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/table"
)

const (
	CACHED_KEY_GUILD_WAR_INFO = "guild_war_info"

	EXPIRE_TIME = base.SecondsPerDay
)

type GuildWarInfoModel struct {
	Creator int
	GId     int
	WarId   int
}

func (this *GuildWarInfoModel) CachedKey() string {
	return fmt.Sprintf("%s.%d.%s", CACHED_KEY_GUILD_WAR_INFO, this.WarId, FormatGuildId(this.Creator, this.GId))
}

func (this *GuildWarInfoModel) Validate() (int, error) {
	if this.Creator <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("guild creator is invalid")
	}

	if this.GId <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("guild id is invalid")
	}

	if this.WarId <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("guild war id is invalid")
	}

	return 0, nil
}

func (this *GuildWarInfoModel) GetWarInfo() (*table.TblGuildWarInfo, error) {
	_, err := this.Validate()
	if err != nil {
		return nil, err
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return nil, custom_errors.New("redisMgr is nil")
	}

	mapKV, err := redisMgr.HashGet(redisKey)
	if err != nil {
		return nil, err
	}

	if len(mapKV) <= 0 {
		return nil, nil
	}

	record := new(table.TblGuildWarInfo)
	err = redistool.ConvertRedisHashToObj(mapKV, record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (this *GuildWarInfoModel) UpdateWarInfo(record *table.TblGuildWarInfo) (int, error) {
	retCode, err := this.Validate()
	if err != nil {
		return retCode, err
	}

	if record == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("record is empty")
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisMgr is nil")
	}

	mapKV, err := redistool.ConvertObjToRedisHash(record)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	err = redisMgr.HashSet(redisKey, mapKV)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	redisMgr.Expire(redisKey, EXPIRE_TIME)

	return 0, nil
}

func GetMultiGuildWarInfo(warId int, guildIdSlice ...string) ([]*table.TblGuildWarInfo, error) {
	if warId <= 0 {
		return nil, custom_errors.New("guild war id is invalid")
	}

	guildWarInfoModel := GuildWarInfoModel{}
	guildWarInfoModel.WarId = warId

	keys := make([]string, 0, len(guildIdSlice))
	var creator, gid int
	var ok bool
	for _, guildId := range guildIdSlice {
		creator, gid, ok = ConvertGuildIdToUinAndId(guildId)
		if !ok {
			continue
		}

		guildWarInfoModel.Creator = creator
		guildWarInfoModel.GId = gid
		keys = append(keys, guildWarInfoModel.CachedKey())
	}

	groupKeys, err := GetKeysGroupByClusterRedis(keys...)
	if err != nil {
		return nil, err
	}

	warInfoSlice := make([]*table.TblGuildWarInfo, 0, len(keys))
	for _, single := range groupKeys {
		redisPipeLine := redistool.NewRedisPipeLine()
		for _, key := range single.Keys {
			redisPipeLine.Append(redistool.REDIS_CONTAINER_HASH, "HGETALL", key)
		}

		results, err := redisPipeLine.Run(single.RedisMgr)
		if err != nil {
			return nil, err
		}

		for _, res := range results {
			hashV, err := res.Hash()
			if err != nil {
				return nil, err
			}

			guildWarInfo := new(table.TblGuildWarInfo)
			err = redistool.ConvertRedisHashToObj(hashV, guildWarInfo)
			if err != nil {
				return nil, err
			}

			warInfoSlice = append(warInfoSlice, guildWarInfo)
		}
	}

	return warInfoSlice, nil
}
