package handler

import (
	"sailcraft/fleetsvr_main/model"
)

func GetGuildWarPoints(warId, creator, gid int) (int, error) {
	guildWarPointsModel := model.GuildWarPointsModel{WarId: warId}
	return guildWarPointsModel.GetGuildWarPoints(creator, gid)
}

func IncreaseGuildWarPoints(warId, creator, gid, pointsIncr int) (int, error) {
	guildWarPointsModel := model.GuildWarPointsModel{WarId: warId}
	return guildWarPointsModel.IncrGuildWarPoints(creator, gid, pointsIncr)
}
