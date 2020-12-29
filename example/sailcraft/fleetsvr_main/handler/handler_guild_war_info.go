package handler

import (
	"sailcraft/fleetsvr_main/model"
	"sailcraft/fleetsvr_main/table"
)

func GetGuildWarInfo(creator, gid, warId int) (*table.TblGuildWarInfo, error) {
	guildWarInfoModel := model.GuildWarInfoModel{Creator: creator, GId: gid, WarId: warId}
	warInfo, err := guildWarInfoModel.GetWarInfo()
	if err != nil {
		return nil, err
	}

	if warInfo == nil {
		warInfo = new(table.TblGuildWarInfo)
		warInfo.LoseTimes = 0
		warInfo.Score = 0
		warInfo.TotalTimes = 0
		warInfo.UserCount = 0
		warInfo.WinTimes = 0
		warInfo.Creator = creator
		warInfo.GuildId = gid
	}

	return warInfo, nil
}

func UpdateGuildWarInfo(creator, gid, warId int, record *table.TblGuildWarInfo) (int, error) {
	guildWarInfoModel := model.GuildWarInfoModel{Creator: creator, GId: gid, WarId: warId}
	return guildWarInfoModel.UpdateWarInfo(record)
}

func GetMultiGuildWarInfo(warId int, guildId ...string) ([]*table.TblGuildWarInfo, error) {
	return model.GetMultiGuildWarInfo(warId, guildId...)
}
