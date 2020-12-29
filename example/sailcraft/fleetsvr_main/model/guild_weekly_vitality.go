package model

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/table"
	"strconv"
)

const (
	TABLE_NAME_GUILD_WEEKLY_VITALITY = "guild.weekly.vitality"
	CACHED_KEY_GUILD_WEEKLY_VITALITY = "cache"

	FIELD_GUILD_WEEKLY_VITALITY_FRESH_TIME = "fresh_time"
)

type GuildWeeklyVitalityModel struct {
	Creator int
	Id      int
}

func (this *GuildWeeklyVitalityModel) CachedKey() string {
	return fmt.Sprintf("%s.%s.%s", TABLE_NAME_GUILD_WEEKLY_VITALITY, CACHED_KEY_GUILD_WEEKLY_VITALITY, FormatGuildId(this.Creator, this.Id))
}

func (this *GuildWeeklyVitalityModel) GetVitalityOfAll() (*table.TblGuildWeeklyVitality, error) {
	if this.Creator <= 0 {
		return nil, custom_errors.New("guild creator is invalid")
	}

	if this.Id <= 0 {
		return nil, custom_errors.New("guild id is invalid")
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

	record := new(table.TblGuildWeeklyVitality)
	record.MemberVitality = make(map[int]int)

	for k, v := range mapKV {
		uin, err := strconv.Atoi(k)
		if err != nil {
			if k == FIELD_GUILD_WEEKLY_VITALITY_FRESH_TIME {
				t, err := strconv.Atoi(v)
				if err != nil {
					continue
				}

				record.FreshTime = t
			}

			continue
		}

		vitality, err := strconv.Atoi(v)
		if err != nil {
			base.GLog.Error(err)
			continue
		}

		record.MemberVitality[uin] = vitality
	}

	return record, nil
}

func (this *GuildWeeklyVitalityModel) GetFreshTime() (int, error) {
	if this.Creator <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("guild creator is invalid")
	}

	if this.Id <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("guild id is invalid")
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisMgr is nil")
	}

	v, err := redisMgr.HashGetFields(redisKey, FIELD_GUILD_WEEKLY_VITALITY_FRESH_TIME)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if len(v) > 0 {
		if v[0] != "" {
			freshTime, err := strconv.Atoi(v[0])
			if err != nil {
				base.GLog.Error(err)
				return 0, nil
			}

			return freshTime, nil
		}
	}

	return 0, nil
}

func (this *GuildWeeklyVitalityModel) SetVitalityOfMember(uin, viality int) (int, error) {
	if this.Creator <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("guild creator is invalid")
	}

	if this.Id <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("guild id is invalid")
	}

	if uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisMgr is nil")
	}

	err := redisMgr.HashSet(redisKey, map[string]string{strconv.Itoa(uin): strconv.Itoa(viality)})
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *GuildWeeklyVitalityModel) Clear(record *table.TblGuildWeeklyVitality) (int, error) {
	if this.Creator <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("guild creator is invalid")
	}

	if this.Id <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("guild id is invalid")
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

	err = redisMgr.HashSet(redisKey, map[string]string{FIELD_GUILD_WEEKLY_VITALITY_FRESH_TIME: strconv.Itoa(int(base.GLocalizedTime.SecTimeStamp()))})
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if record != nil {
		record.FreshTime = int(base.GLocalizedTime.SecTimeStamp())
		record.MemberVitality = make(map[int]int)
	}

	return 0, nil
}

func (this *GuildWeeklyVitalityModel) DeleteGuildMemberVitality(uin int) (int, error) {
	if this.Creator <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("guild creator is invalid")
	}

	if this.Id <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("guild id is invalid")
	}

	if uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisMgr is nil")
	}

	_, err := redisMgr.HashDel(redisKey, strconv.Itoa(uin))
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *GuildWeeklyVitalityModel) GetVitality(uin int) (int, error) {
	if this.Creator <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("guild creator is invalid")
	}

	if this.Id <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("guild id is invalid")
	}

	if uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisMgr is nil")
	}

	v, err := redisMgr.HashGetFields(redisKey, strconv.Itoa(uin))
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if len(v) > 0 {
		vitaliy, err := strconv.Atoi(v[0])
		if err != nil {
			base.GLog.Error(err)
			return 0, nil
		}

		return vitaliy, nil
	}

	return 0, nil
}
