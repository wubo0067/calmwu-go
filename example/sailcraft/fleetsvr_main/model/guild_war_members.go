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
	CACHED_KEY_GUILD_WAR_MEMBERS = "guild_war_members"
)

type GuildWarMembersModel struct {
	Creator int
	GId     int
	WarId   int
}

func (this *GuildWarMembersModel) CachedKey() string {
	return fmt.Sprintf("%s.%d.%s", CACHED_KEY_GUILD_WAR_MEMBERS, this.WarId, FormatGuildId(this.Creator, this.GId))
}

func (this *GuildWarMembersModel) Validate() (int, error) {
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

func (this *GuildWarMembersModel) GetGuildWarMemberInfo(uin int) (*table.TblGuildWarMemberInfo, error) {
	_, err := this.Validate()
	if err != nil {
		return nil, err
	}

	if uin <= 0 {
		return nil, custom_errors.New("uin is invalid")
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return nil, custom_errors.New("redisMgr is nil")
	}

	fields, err := redisMgr.HashGetFields(redisKey, strconv.Itoa(uin))
	if err != nil {
		return nil, err
	}

	if len(fields) <= 0 || fields[0] == "" {
		return nil, nil
	}

	record := new(table.TblGuildWarMemberInfo)
	err = json.Unmarshal([]byte(fields[0]), record)
	if err != nil {
		return nil, nil
	}

	return record, nil
}

func (this *GuildWarMembersModel) GetAllWarMemberInfo() (map[int]*table.TblGuildWarMemberInfo, error) {
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

	mapMem := make(map[int]*table.TblGuildWarMemberInfo)

	for k, v := range mapKV {
		uin, err := strconv.Atoi(k)
		if err != nil {
			base.GLog.Error("hash key[%s] is not uin in redis key[%s], error:%s", k, redisKey, err)
			continue
		}

		record := new(table.TblGuildWarMemberInfo)
		err = json.Unmarshal([]byte(v), record)
		if err != nil {
			base.GLog.Error("json unmarshal %s error in redis key[%s]: %s", k, redisKey, err)
			continue
		}

		mapMem[uin] = record
	}

	return mapMem, nil
}

func (this *GuildWarMembersModel) UpdateMember(record *table.TblGuildWarMemberInfo) (int, error) {
	retCode, err := this.Validate()
	if err != nil {
		return retCode, err
	}

	if record == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("record is empty")
	}

	if record.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("record uin is invalid")
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisMgr is nil")
	}

	data, err := json.Marshal(record)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	err = redisMgr.HashSet(redisKey, map[string]string{strconv.Itoa(record.Uin): string(data)})
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	redisMgr.Expire(redisKey, EXPIRE_TIME)

	return 0, nil
}
