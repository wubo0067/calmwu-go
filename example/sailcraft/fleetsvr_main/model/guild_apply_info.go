package model

import (
	"encoding/json"
	"fmt"
	"sailcraft/base"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/table"
	"strconv"
)

const (
	TABLE_NAME_GUILD_APPLY_INFO       = "guild_apply_info"
	GUILD_APPLY_INFO_REDIS_CACHED_KEY = "cache"
)

type GuildApplyInfoModel struct {
	CreatorUin int
	Id         int
}

func (this *GuildApplyInfoModel) CachedKey() string {
	return fmt.Sprintf("%s.%s.%d.%d", TABLE_NAME_GUILD_APPLY_INFO, GUILD_APPLY_INFO_REDIS_CACHED_KEY, this.CreatorUin, this.Id)
}

func (this *GuildApplyInfoModel) CachedListKey() string {
	return fmt.Sprintf("%s.%s.%d.%d.list", TABLE_NAME_GUILD_APPLY_INFO, GUILD_APPLY_INFO_REDIS_CACHED_KEY, this.CreatorUin, this.Id)
}

func (this *GuildApplyInfoModel) AddApplyInfo(record *table.TblGuildApplyInfo) (int, error) {
	if record == nil {
		return 0, custom_errors.New("record is nil")
	}

	if record.ApplyUin <= 0 {
		return 0, custom_errors.New("apply uin is invalid")
	}

	if this.CreatorUin <= 0 {
		return 0, custom_errors.New("creator's uin is invalid")
	}

	if this.Id <= 0 {
		return 0, custom_errors.New("id is invalid")
	}

	data, err := json.Marshal(record)
	if err != nil {
		return 0, err
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return 0, custom_errors.New("redisMgr[%s] is nil", redisKey)
	}

	err = redisMgr.HashSet(redisKey, map[string]string{strconv.Itoa(record.ApplyUin): string(data)})
	if err != nil {
		return 0, err
	}

	redisListKey := this.CachedListKey()
	redisListMgr := GetClusterRedis(redisListKey)
	if redisListMgr == nil {
		return 0, custom_errors.New("redisListMgr[%s] is nil", redisListKey)
	}

	elemCount, err := redisListMgr.ListRPushVariable(redisListKey, strconv.Itoa(record.ApplyUin))
	if err != nil {
		return 0, err
	}

	return elemCount, nil
}

func (this *GuildApplyInfoModel) GetApplyCount() (int, error) {
	if this.CreatorUin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("creator's uin is invalid")
	}

	if this.Id <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("id is invalid")
	}

	redisListKey := this.CachedListKey()
	redisMgr := GetClusterRedis(redisListKey)
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisMgr[%s] is nil", redisListKey)
	}

	len, err := redisMgr.ListLen(redisListKey)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return len, nil
}

func (this *GuildApplyInfoModel) GetApplyInfo(uin int) (*table.TblGuildApplyInfo, error) {
	if this.CreatorUin <= 0 {
		return nil, custom_errors.New("creator's uin is invalid")
	}

	if this.Id <= 0 {
		return nil, custom_errors.New("id is invalid")
	}

	if uin <= 0 {
		return nil, custom_errors.New("uin is invalid")
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return nil, custom_errors.New("redisMgr[%s] is nil", redisKey)
	}

	applies, err := redisMgr.HashGetFields(redisKey, strconv.Itoa(uin))
	if err != nil {
		return nil, err
	}

	if len(applies) > 0 && applies[0] != "" {
		applyInfo := new(table.TblGuildApplyInfo)

		err = json.Unmarshal([]byte(applies[0]), applyInfo)
		if err != nil {
			return nil, custom_errors.New("Unmarshal apply info error![%s]", err)
		}

		return applyInfo, nil
	}

	return nil, nil
}

// 获取所有申请消息
func (this *GuildApplyInfoModel) GetAllApplyInfo() (map[int]*table.TblGuildApplyInfo, error) {
	if this.CreatorUin <= 0 {
		return nil, custom_errors.New("creator's uin is invalid")
	}

	if this.Id <= 0 {
		return nil, custom_errors.New("id is invalid")
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return nil, custom_errors.New("redisMgr[%s] is nil", redisKey)
	}

	mapKV, err := redisMgr.HashGet(redisKey)
	if err != nil {
		return nil, err
	}

	applyMap := make(map[int]*table.TblGuildApplyInfo)
	for _, applyStr := range mapKV {
		applyInfo := new(table.TblGuildApplyInfo)
		err = json.Unmarshal([]byte(applyStr), applyInfo)
		if err != nil {
			base.GLog.Error("Unmarshal apply info error![%s]", err)
			continue
		}

		applyMap[applyInfo.ApplyUin] = applyInfo
	}

	return applyMap, nil
}

// 删除申请消息
func (this *GuildApplyInfoModel) DeleteApplyInfo(uinList ...int) (int, error) {
	if this.CreatorUin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("creator's uin is invalid")
	}

	if this.Id <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("id is invalid")
	}

	if len(uinList) <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("apply uin list is empty")
	}

	delApplyUin := make([]string, 0, len(uinList))
	for _, applyUin := range uinList {
		if applyUin <= 0 {
			base.GLog.Error("apply uin[%d] is invalid", applyUin)
			continue
		}
		delApplyUin = append(delApplyUin, strconv.Itoa(applyUin))
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisMgr[%s] is nil", redisKey)
	}

	_, err := redisMgr.HashDel(redisKey, delApplyUin...)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	redisListKey := this.CachedListKey()
	redisMgr = GetClusterRedis(redisListKey)
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisMgr[%s] is nil", redisKey)
	}

	for _, v := range delApplyUin {
		_, err = redisMgr.ListRem(redisListKey, v, 1)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
	}

	return 0, nil
}

// 删除所有申请消息
func (this *GuildApplyInfoModel) DeleteAllApplyInfo() (int, error) {
	if this.CreatorUin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("creator's uin is invalid")
	}

	if this.Id <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("id is invalid")
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisMgr[%s] is nil", redisKey)
	}

	err := redisMgr.DelKey(redisKey)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	redisListKey := this.CachedListKey()
	redisMgr = GetClusterRedis(redisListKey)
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisMgr[%s] is nil", redisListKey)
	}

	err = redisMgr.DelKey(redisListKey)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *GuildApplyInfoModel) DeleteOldestApplyInfo(count int) (int, error) {
	if this.CreatorUin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("creator's uin is invalid")
	}

	if this.Id <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("id is invalid")
	}

	redisListKey := this.CachedListKey()
	redisMgrForList := GetClusterRedis(redisListKey)

	delFields := make([]string, 0, count)
	for i := 0; i < count; i++ {
		v, err := redisMgrForList.ListLPop(redisListKey)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		delFields = append(delFields, v)
	}

	base.GLog.Debug("Del Fields: %+v", delFields)
	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	_, err := redisMgr.HashDel(redisKey, delFields...)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}
