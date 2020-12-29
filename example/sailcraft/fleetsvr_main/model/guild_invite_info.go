package model

import (
	"encoding/json"
	"fmt"
	"sailcraft/base"
	"sailcraft/dataaccess/redistool"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/table"
)

type GuildInviteInfoModel struct {
	Uin int
}

const (
	TABLE_NAME_GUILD_INVITE_INFO       = "guild_invite_info"
	GUILD_INVITE_INFO_REDIS_CACHED_KEY = "cache"
)

func (this *GuildInviteInfoModel) CachedKey() string {
	return fmt.Sprintf("%s.%s.%d", TABLE_NAME_GUILD_INVITE_INFO, GUILD_INVITE_INFO_REDIS_CACHED_KEY, this.Uin)
}

func (this *GuildInviteInfoModel) CachedListKey() string {
	return fmt.Sprintf("%s.%s.%d.list", TABLE_NAME_GUILD_INVITE_INFO, GUILD_INVITE_INFO_REDIS_CACHED_KEY, this.Uin)
}

func (this *GuildInviteInfoModel) AddGuildInvite(record *table.TblGuildInvite) (int, error) {
	if this.Uin <= 0 {
		return 0, custom_errors.New("uin is invalid")
	}

	if record == nil {
		return 0, custom_errors.New("record is empty")
	}

	data, err := json.Marshal(record)
	if err != nil {
		return 0, err
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return 0, custom_errors.New("redisMgr is nil")
	}

	err = redisMgr.HashSet(redisKey, map[string]string{record.GuildId: string(data)})
	if err != nil {
		return 0, err
	}

	redisListKey := this.CachedListKey()
	redisListMgr := GetClusterRedis(redisListKey)
	if redisListMgr == nil {
		return 0, custom_errors.New("redisListMgr is nil")
	}

	elemCount, err := redisListMgr.ListRPushVariable(redisListKey, record.GuildId)
	if err != nil {
		return 0, err
	}

	return elemCount, nil
}

func (this *GuildInviteInfoModel) GetGuildInviteCount() (int, error) {
	if this.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	redisListKey := this.CachedListKey()
	redisListMgr := GetClusterRedis(redisListKey)
	if redisListMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisListMgr is nil")
	}

	inviteCount, err := redisListMgr.ListLen(redisListKey)
	if err != nil {
		return 0, err
	}

	return inviteCount, nil
}

func (this *GuildInviteInfoModel) GetAllGuildInvite() ([]*table.TblGuildInvite, error) {
	if this.Uin <= 0 {
		return nil, custom_errors.New("uin is invalid")
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

	inviteSlice := make([]*table.TblGuildInvite, 0, len(mapKV))
	for _, inviteStr := range mapKV {
		inviteInfo := new(table.TblGuildInvite)
		err := json.Unmarshal([]byte(inviteStr), inviteInfo)
		if err != nil {
			base.GLog.Error("Unmarshal invite info error![%s]", err)
			continue
		}

		inviteSlice = append(inviteSlice, inviteInfo)
	}

	return inviteSlice, nil
}

func (this *GuildInviteInfoModel) GetGuildInviteInfoByGuildId(guildId string) (*table.TblGuildInvite, error) {
	if this.Uin <= 0 {
		return nil, custom_errors.New("uin is invalid")
	}

	if !ValidGuildId(guildId) {
		return nil, custom_errors.New("guild id format error")
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return nil, custom_errors.New("redisMgr is nil")
	}

	invites, err := redisMgr.HashGetFields(redisKey, guildId)
	if err != nil {
		return nil, err
	}

	if len(invites) > 0 && invites[0] != "" {
		inviteInfo := new(table.TblGuildInvite)

		err = json.Unmarshal([]byte(invites[0]), inviteInfo)
		if err != nil {
			return nil, err
		}

		return inviteInfo, nil
	}

	return nil, nil
}

func (this *GuildInviteInfoModel) DeleteGuildInviteByGuildId(guildIdSlice ...string) (int, error) {
	if this.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	if len(guildIdSlice) <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin list is empty")
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisMgr[%s] is nil", redisKey)
	}

	_, err := redisMgr.HashDel(redisKey, guildIdSlice...)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	redisListKey := this.CachedListKey()
	redisMgr = GetClusterRedis(redisListKey)
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisMgr[%s] is nil", redisKey)
	}

	for _, v := range guildIdSlice {
		_, err = redisMgr.ListRem(redisListKey, v, 1)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
	}

	return 0, nil
}

func (this *GuildInviteInfoModel) DeleteAllInvite() (int, error) {
	if this.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisMgr is nil")
	}

	err := redisMgr.DelKey(redisKey)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	redisListKey := this.CachedListKey()
	redisListMgr := GetClusterRedis(redisListKey)
	if redisListMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisListMgr is nil")
	}

	err = redisListMgr.DelKey(redisListKey)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *GuildInviteInfoModel) DeleteOldestInvitInfo(count int) (int, error) {
	if this.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	redisListKey := this.CachedListKey()
	redisListMgr := GetClusterRedis(redisListKey)
	if redisListMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisListMgr is nil")
	}

	redisPipeLine := redistool.NewRedisPipeLine()
	for i := 0; i < count; i++ {
		redisPipeLine.Append(redistool.REDIS_CONTAINER_LIST, "LPOP", redisListKey)
	}
	results, err := redisPipeLine.Run(redisListMgr)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	delFields := make([]string, 0, count)
	for _, res := range results {
		v, err := res.String()
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		delFields = append(delFields, v)
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisMgr is nil")
	}

	base.GLog.Debug("Delete Fields: %+v", delFields)
	_, err = redisMgr.HashDel(redisKey, delFields...)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}
