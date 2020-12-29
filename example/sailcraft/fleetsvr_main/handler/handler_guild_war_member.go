package handler

import (
	"sailcraft/fleetsvr_main/config"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/model"
	"sailcraft/fleetsvr_main/proto"
	"sailcraft/fleetsvr_main/table"
)

func GetGuildWarMember(creator, gId, warId, uin int) (*table.TblGuildWarMemberInfo, error) {
	guildWarMembersModel := model.GuildWarMembersModel{Creator: creator, GId: gId, WarId: warId}
	warMemberInfo, err := guildWarMembersModel.GetGuildWarMemberInfo(uin)
	if err != nil {
		return nil, err
	}

	if warMemberInfo == nil {
		warMemberInfo = new(table.TblGuildWarMemberInfo)
		warMemberInfo.Uin = uin
		warMemberInfo.Score = 0
		warMemberInfo.GuildId = FormatGuildId(creator, gId)
		warMemberInfo.RestBattleTimes = config.GGlobalConfig.Guild.BattleConfig.PlayerBattleTimes
	}

	return warMemberInfo, nil
}

func UpdateGuildWarMember(creator, gId, warId int, record *table.TblGuildWarMemberInfo) (int, error) {
	guildWarMembersModel := model.GuildWarMembersModel{Creator: creator, GId: gId, WarId: warId}
	return guildWarMembersModel.UpdateMember(record)
}

func GetAllWarMemberOfGuild(creator, gId, warId int) ([]*table.TblGuildWarMemberInfo, error) {
	guildWarMembersModel := model.GuildWarMembersModel{Creator: creator, GId: gId, WarId: warId}
	mapKV, err := guildWarMembersModel.GetAllWarMemberInfo()
	if err != nil {
		return nil, err
	}

	members := make([]*table.TblGuildWarMemberInfo, 0, len(mapKV))
	for _, v := range mapKV {
		members = append(members, v)
	}

	return members, nil
}

func composeProtoGuildMemberBattleInfo(target *proto.ProtoGuildMemberBattleInfo, data *table.TblGuildWarMemberInfo, userInfo *table.TblUserInfo, userHeadFrame *table.TblUserHeadFrame) (int, error) {
	if target == nil || data == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	target.Uin = data.Uin
	target.Score = data.Score
	if userInfo != nil {
		target.UserName = userInfo.UserName
		target.Icon = userInfo.Icon
		target.CountryCode = userInfo.ISOCountryCode
		target.Level = userInfo.Level
	}

	if userHeadFrameIcon, ok := config.GIconFrameConfig.AttrMap[userHeadFrame.CurHeadFrame]; ok {
		target.CurHeadFrame = userHeadFrameIcon.ResourceId
		target.HeadFrameProtypeId = userHeadFrame.CurHeadFrame
		target.HeadId = userHeadFrame.HeadId
		target.HeadType = userHeadFrame.HeadType
	} else {
		target.CurHeadFrame = ""
	}

	target.BattleRecords = make([]*proto.ProtoGuildMemberBattleRecord, 0, len(data.BattleRecordList))
	for _, battleRecord := range data.BattleRecordList {
		protoBattleRecord := new(proto.ProtoGuildMemberBattleRecord)
		protoBattleRecord.BattleResult = battleRecord.BattleResult
		protoBattleRecord.ScoreIncr = battleRecord.ScoreIncr
		target.BattleRecords = append(target.BattleRecords, protoBattleRecord)
	}

	return 0, nil
}

func composeProtoGuildMemberBattleInfoList(memberSlice ...*table.TblGuildWarMemberInfo) ([]*proto.ProtoGuildMemberBattleInfo, error) {
	uinList := make([]int, 0, len(memberSlice))
	for _, member := range memberSlice {
		uinList = append(uinList, member.Uin)
	}

	allUsers, err := GetMultiUserInfo(uinList...)
	if err != nil {
		return nil, err
	}

	userInfoMap := make(map[int]*table.TblUserInfo)
	for _, userInfo := range allUsers {
		userInfoMap[userInfo.Uin] = userInfo
	}

	userHeadFrameMap, err := GetMultiUserHeadFrame(uinList...)
	if err != nil {
		return nil, err
	}

	protoGuildMemberBattleInfoList := make([]*proto.ProtoGuildMemberBattleInfo, 0, len(memberSlice))
	for _, member := range memberSlice {
		protoGuildMemberBattleInfo := new(proto.ProtoGuildMemberBattleInfo)
		composeProtoGuildMemberBattleInfo(protoGuildMemberBattleInfo, member, userInfoMap[member.Uin], userHeadFrameMap[member.Uin])
		if err != nil {
			return nil, err
		}

		protoGuildMemberBattleInfoList = append(protoGuildMemberBattleInfoList, protoGuildMemberBattleInfo)
	}

	return protoGuildMemberBattleInfoList, nil

}
