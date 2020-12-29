package handler

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/dataaccess/redistool"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/model"
	"sailcraft/fleetsvr_main/table"
)

const (
	GUILD_WEEKLY_VITALITY_LOCKER = "guild.weekly_vitality.locker"
)

func GetGuildWeeklyVitalityOfAll(creator, id int) (*table.TblGuildWeeklyVitality, error) {
	guildWeeklyVitalityModel := model.GuildWeeklyVitalityModel{Creator: creator, Id: id}
	weeklyVitality, err := guildWeeklyVitalityModel.GetVitalityOfAll()
	if err != nil {
		return nil, err
	}

	if base.GLocalizedTime.IsCurrentWeek(int64(weeklyVitality.FreshTime)) {
		err = clearGuildWeeklyVitality(creator, id, weeklyVitality)
		if err != nil {
			return nil, err
		}
	}

	return weeklyVitality, nil
}

func GetGuildWeeklyVitality(creator, id, uin int) (int, error) {
	guildWeeklyVitalityModel := model.GuildWeeklyVitalityModel{Creator: creator, Id: id}
	freshTime, err := guildWeeklyVitalityModel.GetFreshTime()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if base.GLocalizedTime.IsCurrentWeek(int64(freshTime)) {
		err = clearGuildWeeklyVitality(creator, id, nil)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
	}

	return guildWeeklyVitalityModel.GetVitality(uin)
}

func SetGuildWeeklyVitalityOfMember(creator, id, uin, viality int) (int, error) {
	guildWeeklyVitalityModel := model.GuildWeeklyVitalityModel{Creator: creator, Id: id}
	freshTime, err := guildWeeklyVitalityModel.GetFreshTime()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if base.GLocalizedTime.IsCurrentWeek(int64(freshTime)) {
		err = clearGuildWeeklyVitality(creator, id, nil)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
	}

	return guildWeeklyVitalityModel.SetVitalityOfMember(uin, viality)
}

func DeleteGuildWeeklyVitalityOfMember(creator, id, uin int) (int, error) {
	guildWeeklyVitalityModel := model.GuildWeeklyVitalityModel{Creator: creator, Id: id}
	return guildWeeklyVitalityModel.DeleteGuildMemberVitality(uin)
}

func clearGuildWeeklyVitality(creator, id int, weeklyVitality *table.TblGuildWeeklyVitality) error {
	locker, err := LockGuildWeeklyVitality(creator, id)
	if err != nil {
		return err
	}
	defer UnlockGuildWeeklyVitality(creator, id, locker)

	guildWeeklyVitalityModel := model.GuildWeeklyVitalityModel{Creator: creator, Id: id}
	freshTime, err := guildWeeklyVitalityModel.GetFreshTime()
	if base.GLocalizedTime.IsCurrentWeek(int64(freshTime)) {
		_, err := guildWeeklyVitalityModel.Clear(weeklyVitality)
		if err != nil {
			return err
		}
	}

	return nil
}

func LockGuildWeeklyVitality(creator int, gId int) (string, error) {
	k := fmt.Sprintf("%s.%d.%d", GUILD_WEEKLY_VITALITY_LOCKER, creator, gId)
	return redistool.SpinLockWithFingerPoint(k, 0)
}

func UnlockGuildWeeklyVitality(creator int, gId int, value string) {
	k := fmt.Sprintf("%s.%d.%d", GUILD_WEEKLY_VITALITY_LOCKER, creator, gId)
	err := redistool.UnLock(k, value)
	if err != nil {
		base.GLog.Error("unlock guild weekly vitality(creatorUin:%d, gId:%d, value:%s) error[%v]", creator, gId, value, err)
	}
}
