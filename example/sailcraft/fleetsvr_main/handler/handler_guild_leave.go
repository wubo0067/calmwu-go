package handler

import (
	"sailcraft/base"
	"sailcraft/fleetsvr_main/config"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/model"
	"sailcraft/fleetsvr_main/table"
)

func GetGuildLeaveInfo(uin int) (*table.TblGuildLeaveInfo, error) {
	guildLeaveModel := model.GuildLeaveInfoModel{Uin: uin}
	return guildLeaveModel.GetGuildLeaveInfo()
}

func SetGuildLeaveInfo(uin int, record *table.TblGuildLeaveInfo) (int, error) {
	guildLeaveModel := model.GuildLeaveInfoModel{Uin: uin}
	return guildLeaveModel.SetGuildLeaveInfo(record)
}

func ValidateGuildLeaveTime(leaveInfo *table.TblGuildLeaveInfo) (int, error) {
	if leaveInfo != nil {
		diff := int(base.GLocalizedTime.SecTimeStamp()) - leaveInfo.LeaveTime
		restTime := config.GGlobalConfig.Guild.JoinGuildColdTime - diff
		if restTime > 0 {
			return restTime, custom_errors.New("join guild cold time is not reach")
		}
	}

	return 0, nil
}
