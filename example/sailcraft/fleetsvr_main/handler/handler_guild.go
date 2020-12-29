package handler

import (
	"encoding/json"
	"fmt"
	"sailcraft/base"
	"sailcraft/dataaccess/redistool"
	"sailcraft/fleetsvr_main/config"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/handlerbase"
	"sailcraft/fleetsvr_main/model"
	"sailcraft/fleetsvr_main/proto"
	"sailcraft/fleetsvr_main/table"
	"unicode/utf8"
)

const (
	GUILD_JOIN_TYPE_ALLOW_ANYONE = 0
	GUILD_JOIN_TYPE_NEED_AUDIT   = 1
	GUILD_JOIN_TYPE_NOT_ALLOWED  = 2

	GUILD_APPLY_OPERATION_ALLOW = 0
	GUILD_APPLY_OPERATION_DENY  = 1

	GUILD_INVITE_OPERATION_AGREE  = 0
	GUILD_INVITE_OPERATION_REFUSE = 1

	GUILD_LOCKER      = "guild.locker"
	GUILD_USER_LOCKER = "guild.user.locker"
)

type GuildHandler struct {
	handlerbase.WebHandler
}

var GuildPostSort = map[string]int{GUILD_POST_CHAIRMAN: 0, GUILD_POST_VICE_CHAIRMAN: 1, GUILD_POST_ELDER: 2, GUILD_POST_MEMBER: 3}
var GuildPostSort2Post = map[int]string{0: GUILD_POST_CHAIRMAN, 1: GUILD_POST_VICE_CHAIRMAN, 2: GUILD_POST_ELDER, 3: GUILD_POST_MEMBER}

func (this *GuildHandler) Create() (int, error) {
	var reqParams proto.ProtoCreateGuildRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	// 1. 检测创建公会信息是否合法
	retCode, err := ValidCreateGuildInfo(&reqParams)
	if err != nil {
		return retCode, err
	}

	var userInfo table.TblUserInfo
	retCode, err = GetUserInfo(this.Request.Uin, &userInfo)
	if err != nil {
		return retCode, err
	}

	if ValidGuildId(userInfo.GuildID) {
		return errorcode.ERROR_CODE_GUILD_ALREADY_IN_GUILD, custom_errors.New("user already has guild.")
	}

	// 2. 检测玩家当前是否可以创建公会
	realCostResources, retCode, err := ValidUserCreateGuildPermissionAndCost(&userInfo)
	if err != nil {
		return retCode, err
	}

	performId, err := model.GenerateGuildPerformId(reqParams.ServerId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	tblGuild := NewGuild(this.Request.Uin, reqParams.GuildInfo.Name)
	tblGuild.PerformId = performId
	SetGuildModifiableInfo(tblGuild, &reqParams.GuildInfo.ProtoGuildModifiableInfo)

	// 添加公会成员信息以及更新玩家公会信息
	memberInfo, retCode, err := AddNewGuildMember(tblGuild, &userInfo, GUILD_POST_CHAIRMAN)
	if err != nil {
		return retCode, err
	}

	leagueInfo, err := GetLeagueInfo(this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	// 添加公会护卫舰
	_, err = CreateFrigateShipInfo(tblGuild.Id, tblGuild.Creator)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	var responseData proto.ProtoCreateGuildResponse

	// 消耗
	var realCost config.ResourcesAttr
	realCost.ResourceItems = realCostResources
	ResourcesConfigToProto(&realCost, &responseData.Cost)
	// 公会信息
	retCode, err = composeProtoGuildInfo(&responseData.GuildInfo, tblGuild, &userInfo, 0)
	if err != nil {
		return retCode, err
	}
	// 公会成员
	weeklyVitality, err := GetGuildWeeklyVitality(tblGuild.Creator, tblGuild.Id, this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	tblUserheadInfo, err := GetUserHeadFrame(userInfo.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	protoGuildMemberInfo := new(proto.ProtoGuildMemberInfo)
	retCode, err = composeProtoGuildMemberInfo(protoGuildMemberInfo, memberInfo, &userInfo, leagueInfo, weeklyVitality, tblUserheadInfo)
	if err != nil {
		return retCode, err
	}
	responseData.GuildMembers = append(responseData.GuildMembers, protoGuildMemberInfo)

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *GuildHandler) ModifyGuildInfo() (int, error) {
	var reqParams proto.ProtoModifyGuildInfoRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	// 检测公会信息格式是否合法
	retCode, err := ValidGuildModifiableInfo(&reqParams.ProtoGuildModifiableInfo)
	if err != nil {
		return retCode, err
	}

	var userInfo table.TblUserInfo
	retCode, err = GetUserInfo(this.Request.Uin, &userInfo)
	if err != nil {
		return retCode, err
	}

	// 玩家不在公会中
	creatorUin, gId, correct := ConvertGuildIdToUinAndId(userInfo.GuildID)
	if !correct {
		return errorcode.ERROR_CODE_GUILD_NOT_IN_GUILD, custom_errors.New("not in guild")
	}

	locker, err := LockGuild(creatorUin, gId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	defer UnlockGuild(creatorUin, gId, locker)

	// 检测玩家是否具有修改权限
	memberInfo, err := GetGuildMemberInfo(creatorUin, gId, this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if memberInfo == nil {
		return errorcode.ERROR_CODE_GUILD_NOT_IN_GUILD, custom_errors.New("member is not in guild")
	}

	if memberInfo.Post != GUILD_POST_CHAIRMAN && memberInfo.Post != GUILD_POST_VICE_CHAIRMAN {
		return errorcode.ERROR_CODE_GUILD_PERMISSION_DENIED, custom_errors.New("can not modify guild info due to inadequate permissions")
	}

	guildModel := model.GuildInfoModel{Creator: creatorUin, Id: gId}
	tblGuild, err := guildModel.GetGuildInfo()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if tblGuild == nil {
		return errorcode.ERROR_CODE_GUILD_ID_NOT_EXIST, custom_errors.New("guild id (%s) not exist in database", userInfo.GuildID)
	}

	tblGuild.CondLeagueLevel = reqParams.CondLeagueLevel
	tblGuild.JoinType = reqParams.JoinType
	tblGuild.Description = reqParams.Description
	tblGuild.Symbol = reqParams.Symbol
	retCode, err = guildModel.UpdateGuildInfo(tblGuild)
	if err != nil {
		return retCode, err
	}

	vitality, err := GetGuildVitality(FormatGuildId(tblGuild.Creator, tblGuild.Id))
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	var responseData proto.ProtoModifyGuildInfoResponse
	retCode, err = composeProtoGuildInfo(&responseData.ProtoGuildInfo, tblGuild, nil, vitality)
	if err != nil {
		return retCode, err
	}
	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *GuildHandler) SimpleInfoByUin() (int, error) {
	var reqParams proto.ProtoGetSimpleGuildInfoByUinRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if reqParams.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	var userInfo table.TblUserInfo
	retCode, err := GetUserInfo(reqParams.Uin, &userInfo)
	if err != nil {
		return retCode, err
	}

	creator, id, ok := ConvertGuildIdToUinAndId(userInfo.GuildID)
	if !ok {
		return errorcode.ERROR_CODE_GUILD_NOT_IN_GUILD, custom_errors.New("guild id format error")
	}

	guildInfoModel := model.GuildInfoModel{Creator: creator, Id: id}
	guildInfo, err := guildInfoModel.GetGuildInfo()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if guildInfo == nil {
		return errorcode.ERROR_CODE_GUILD_ID_NOT_EXIST, custom_errors.New("guild not exists")
	}

	vitality, err := GetGuildVitality(userInfo.GuildID)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	protoGuildInfo := new(proto.ProtoGuildInfo)
	retCode, err = composeProtoGuildInfo(protoGuildInfo, guildInfo, nil, vitality)
	if err != nil {
		return retCode, err
	}

	guildMemberInfo, err := GetGuildMemberInfo(creator, id, reqParams.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	protoGuildMemberInfo := new(proto.ProtoGuildMemberInfo)
	composeProtoGuildMemberInfo(protoGuildMemberInfo, guildMemberInfo, nil, nil, 0, nil)

	var responseData proto.ProtoGetSimpleGuildInfoByUinResponse
	responseData.GuildInfo = protoGuildInfo
	responseData.GuildMemberInfo = protoGuildMemberInfo

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *GuildHandler) GuildInfoWithMembers() (int, error) {
	var reqParams proto.ProtoGetGuildInfoRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	creator, id, ok := ConvertGuildIdToUinAndId(reqParams.GuildId)
	if !ok {
		return errorcode.ERROR_CODE_GUILD_ID_FORMAT_WRONG, custom_errors.New("guild id format is wrong")
	}

	guildInfoModel := model.GuildInfoModel{Creator: creator, Id: id}
	guildInfo, err := guildInfoModel.GetGuildInfo()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if guildInfo == nil {
		return errorcode.ERROR_CODE_GUILD_ID_NOT_EXIST, custom_errors.New("guild is not exist")
	}

	oldChairman := new(table.TblUserInfo)
	_, err = GetUserInfo(guildInfo.Chairman, oldChairman)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	nextChairman, err := CheckChairmanOfflineTime(guildInfo, oldChairman)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	allMembers, err := GetAllGuildMemberInfo(creator, id)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	protoGuildMemberList, chairman, err := composeProtoGuildMemberList(creator, id, allMembers...)

	var responseData proto.ProtoGetGuildInfoResponse
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	responseData.GuildMembers = protoGuildMemberList

	vitality, err := GetGuildVitality(reqParams.GuildId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	retCode, err := composeProtoGuildInfo(&responseData.GuildInfo, guildInfo, chairman, vitality)
	if err != nil {
		return retCode, err
	}

	responseData.MessageList = make([]*proto.ProtoMessageInfo, 0)
	if nextChairman != nil {
		protoMessageInfo, err := AddGuildChairmanTrasferMessage(reqParams.GuildId, this.Request.Uin, oldChairman, nextChairman)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
		responseData.MessageList = append(responseData.MessageList, protoMessageInfo)
	}

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *GuildHandler) SimpleGuildInfoWithMembers() (int, error) {
	var reqParams proto.ProtoGetSimpleGuildInfoWithMembersRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	creator, id, ok := ConvertGuildIdToUinAndId(reqParams.GuildId)
	if !ok {
		return errorcode.ERROR_CODE_GUILD_ID_FORMAT_WRONG, custom_errors.New("guild id format is wrong")
	}

	guildInfoModel := model.GuildInfoModel{Creator: creator, Id: id}
	guildInfo, err := guildInfoModel.GetGuildInfo()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if guildInfo == nil {
		return errorcode.ERROR_CODE_GUILD_ID_NOT_EXIST, custom_errors.New("guild is not exist")
	}

	guildMemberInfoModel := model.GuildMemberInfoModel{Creator: creator, GId: id}
	allMembers, err := guildMemberInfoModel.GetAllGuildMember()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	var responseData proto.ProtoGetSimpleGuildInfoWithMembersResponse
	responseData.GuildMembers = make([]*proto.ProtoGuildMemberInfo, 0, len(allMembers))
	for _, member := range allMembers {
		protoMemberInfo := new(proto.ProtoGuildMemberInfo)
		retCode, err := composeProtoGuildMemberInfo(protoMemberInfo, member, nil, nil, 0, nil)
		if err != nil {
			return retCode, err
		}

		responseData.GuildMembers = append(responseData.GuildMembers, protoMemberInfo)
	}

	retCode, err := composeProtoGuildInfoWithNoCompletion(&responseData.GuildInfo, guildInfo)
	if err != nil {
		return retCode, err
	}

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *GuildHandler) SingleGuildInfo() (int, error) {
	var reqParams proto.ProtoQuerySingleGuildInfoRequest

	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	creator, id, ok := ConvertGuildIdToUinAndId(reqParams.GuildId)

	if !ok {
		return errorcode.ERROR_CODE_GUILD_ID_FORMAT_WRONG, custom_errors.New("guild id format error")
	}

	guildInfoModel := model.GuildInfoModel{Creator: creator, Id: id}
	guildInfo, err := guildInfoModel.GetGuildInfo()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if guildInfo == nil {
		return errorcode.ERROR_CODE_GUILD_ID_NOT_EXIST, custom_errors.New("guild[%s] not exist", reqParams.GuildId)
	}

	vitality, err := GetGuildVitality(reqParams.GuildId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	var responseData proto.ProtoQuerySingleGuildInfoResponse
	responseData.GuildInfo = new(proto.ProtoGuildInfo)
	retCode, err := composeProtoGuildInfo(responseData.GuildInfo, guildInfo, nil, vitality)
	if err != nil {
		return retCode, err
	}

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *GuildHandler) MultiGuildInfo() (int, error) {
	var reqParams proto.ProtoQueryMultiGuildInfoRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	tblGuildInfoList, err := model.GetMultiGuildInfo(reqParams.GuildIdList...)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	uinList := make([]int, 0, len(tblGuildInfoList))
	for _, tblGuildInfo := range tblGuildInfoList {
		uinList = append(uinList, tblGuildInfo.Chairman)
	}

	allUsers, err := GetMultiUserInfo(uinList...)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	userMap := make(map[int]*table.TblUserInfo)
	for _, user := range allUsers {
		userMap[user.Uin] = user
	}

	var responseData proto.ProtoQueryMultiGuildInfoResponse
	responseData.GuildInfoList = make([]*proto.ProtoGuildInfo, 0, len(tblGuildInfoList))
	for _, tblGuildInfo := range tblGuildInfoList {
		protoGuildInfo := new(proto.ProtoGuildInfo)
		vitality, err := GetGuildVitality(FormatGuildId(tblGuildInfo.Creator, tblGuildInfo.Id))
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		retCode, err := composeProtoGuildInfo(protoGuildInfo, tblGuildInfo, userMap[tblGuildInfo.Chairman], vitality)
		if err != nil {
			return retCode, err
		}

		responseData.GuildInfoList = append(responseData.GuildInfoList, protoGuildInfo)
	}

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *GuildHandler) Join() (int, error) {
	var reqParams proto.ProtoJoinGuildRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	creatorUin, id, correct := ConvertGuildIdToUinAndId(reqParams.GuildId)
	if !correct || creatorUin <= 0 || id <= 0 {
		return errorcode.ERROR_CODE_GUILD_ID_FORMAT_WRONG, custom_errors.New("guild id format error[%s]", reqParams.GuildId)
	}

	// 判断玩家距离上一次退出公会时间是否大于冷却时间
	leaveInfo, err := GetGuildLeaveInfo(this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	restTime, err := ValidateGuildLeaveTime(leaveInfo)
	if err != nil {
		var responseData proto.ProtoJoinGuildWaitForLeaveTimeResponse
		responseData.RestWaitTime = restTime
		this.Response.ResData.Params = responseData
		return errorcode.ERROR_CODE_GUILD_JOIN_NEED_WAIT, err
	}

	var userInfo table.TblUserInfo
	retCode, err := GetUserInfo(this.Request.Uin, &userInfo)
	if err != nil {
		return retCode, err
	}

	retCode, err = ValidUserJoinGuildPermission(&userInfo, creatorUin, id)
	if err != nil {
		return retCode, err
	}

	// 分布式锁，防止多个玩家同时修改公会相关信息
	locker, err := LockGuild(creatorUin, id)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	defer UnlockGuild(creatorUin, id, locker)

	guildModel := model.GuildInfoModel{Creator: creatorUin, Id: id}
	guildInfo, err := guildModel.GetGuildInfo()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if guildInfo == nil {
		return errorcode.ERROR_CODE_GUILD_ID_NOT_EXIST, custom_errors.New("guild not exist")
	}

	retCode, err = ValidGuildJoinStatus(reqParams.GuildId, guildInfo, false)
	if err != nil {
		return retCode, err
	}

	if guildInfo.JoinType != GUILD_JOIN_TYPE_ALLOW_ANYONE {
		return errorcode.ERROR_CODE_GUILD_JOIN_TYPE_NOT_ALLOW_ANYONE, custom_errors.New("guild join type[%d] is not ALLOW_ANYONE(0)", guildInfo.JoinType)
	}

	// 联赛等级检测
	leagueInfo, err := GetLeagueInfo(this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if leagueInfo.CurrentLeagueLevel < guildInfo.CondLeagueLevel {
		return errorcode.ERROR_CODE_GUILD_LEAGUE_LEVEL_NOT_REACH, custom_errors.New("user[%d] league level[%d] not reach condition[%d]", this.Request.Uin, leagueInfo.CurrentLeagueLevel, guildInfo.CondLeagueLevel)
	}

	newGuildMemberInfo, retCode, err := AddNewGuildMember(guildInfo, &userInfo, GUILD_POST_MEMBER)
	if err != nil {
		return retCode, err
	}

	weeklyVitality, err := GetGuildWeeklyVitality(creatorUin, id, userInfo.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	tblUserheadInfo, err := GetUserHeadFrame(userInfo.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	var responseData proto.ProtoJoinGuildSuccessResponse
	responseData.GuildMembers = make([]*proto.ProtoGuildMemberInfo, 0, 1)
	protoGuildMemberInfo := new(proto.ProtoGuildMemberInfo)
	composeProtoGuildMemberInfo(protoGuildMemberInfo, newGuildMemberInfo, &userInfo, leagueInfo, weeklyVitality, tblUserheadInfo)
	responseData.GuildMembers = append(responseData.GuildMembers, protoGuildMemberInfo)

	vitality, err := GetGuildVitality(reqParams.GuildId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	var chairman table.TblUserInfo
	retCode, err = GetUserInfo(guildInfo.Chairman, &chairman)
	if err != nil {
		return retCode, err
	}

	retCode, err = composeProtoGuildInfo(&responseData.GuildInfo, guildInfo, &chairman, vitality)
	if err != nil {
		return retCode, err
	}

	protoMessageInfo, err := AddMemeberCountChangedMessage(reqParams.GuildId, this.Request.Uin, &userInfo, MEMBER_COUNT_OPERATION_JOIN)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	responseData.MessageList = append(responseData.MessageList, protoMessageInfo)

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *GuildHandler) Leave() (int, error) {
	var userInfo table.TblUserInfo
	retCode, err := GetUserInfo(this.Request.Uin, &userInfo)
	if err != nil {
		return retCode, err
	}

	creatorUin, gId, ok := ConvertGuildIdToUinAndId(userInfo.GuildID)
	if !ok {
		return errorcode.ERROR_CODE_GUILD_NOT_IN_GUILD, custom_errors.New("user[%d] is not in guild", userInfo.Uin)
	}

	locker, err := LockGuild(creatorUin, gId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	defer UnlockGuild(creatorUin, gId, locker)

	guildInfoModel := model.GuildInfoModel{Creator: creatorUin, Id: gId}
	guildInfo, err := guildInfoModel.GetGuildInfo()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	delMemberInfo, retCode, err := DeleteGuildMemberByUin(guildInfo, userInfo.Uin)
	if err != nil {
		return retCode, err
	}

	var responseData proto.ProtoLeaveGuildResponse
	if guildInfo.MemberCount <= 0 {
		responseData.DeleteGuild = 1

		responseData.GuildInfo = new(proto.ProtoGuildInfo)
		retCode, err = composeProtoGuildInfo(responseData.GuildInfo, guildInfo, nil, 0)
		if err != nil {
			return retCode, err
		}
	} else {
		responseData.DeleteGuild = 0

		guildVitality, err := GetGuildVitality(userInfo.GuildID)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		var chairman table.TblUserInfo
		retCode, err = GetUserInfo(guildInfo.Chairman, &chairman)
		if err != nil {
			return retCode, err
		}

		responseData.GuildInfo = new(proto.ProtoGuildInfo)
		retCode, err = composeProtoGuildInfo(responseData.GuildInfo, guildInfo, &chairman, guildVitality)
		if err != nil {
			return retCode, err
		}
	}

	leagueInfo, err := GetLeagueInfo(this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	tblUserheadInfo, err := GetUserHeadFrame(this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	protoGuildMemberInfo := new(proto.ProtoGuildMemberInfo)
	composeProtoGuildMemberInfo(protoGuildMemberInfo, delMemberInfo, &userInfo, leagueInfo, 0, tblUserheadInfo)
	responseData.GuildMembers = append(responseData.GuildMembers, protoGuildMemberInfo)

	protoMessageInfo, err := AddMemeberCountChangedMessage(userInfo.GuildID, this.Request.Uin, &userInfo, MEMBER_COUNT_OPERATION_LEAVE)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	responseData.MessageList = append(responseData.MessageList, protoMessageInfo)

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *GuildHandler) KickOutOfGuild() (int, error) {
	var reqParams proto.ProtoKickOutOfGuildRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if reqParams.MemberUin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("member uin is invalid[%d]", reqParams.MemberUin)
	}

	var userInfo table.TblUserInfo
	_, err = GetUserInfo(this.Request.Uin, &userInfo)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	creator, id, ok := ConvertGuildIdToUinAndId(userInfo.GuildID)
	if !ok {
		return errorcode.ERROR_CODE_GUILD_PERMISSION_DENIED, custom_errors.New("user is not in guild")
	}

	locker, err := LockGuild(creator, id)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	defer UnlockGuild(creator, id, locker)

	userMemberInfo, err := GetGuildMemberInfo(creator, id, this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if userMemberInfo == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("data error: guild of 'guild_id' in table 'user_info' doesn't container user[%d]", this.Request.Uin)
	}

	targetMemberInfo, err := GetGuildMemberInfo(creator, id, reqParams.MemberUin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	if targetMemberInfo == nil {
		return errorcode.ERROR_CODE_GUILD_NOT_IN_THIS_GUILD, custom_errors.New("target player is not in this guild")
	}

	userPostSort := GuildPostSort[userMemberInfo.Post]
	if userPostSort >= GuildPostSort[GUILD_POST_ELDER] || userPostSort >= GuildPostSort[targetMemberInfo.Post] {
		return errorcode.ERROR_CODE_GUILD_PERMISSION_DENIED, custom_errors.New("has not permission to kick %s out of guild", targetMemberInfo.Post)
	}

	guildInfoModel := model.GuildInfoModel{Creator: creator, Id: id}
	guildInfo, err := guildInfoModel.GetGuildInfo()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	if guildInfo == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("data error: guild of 'guild_id(%s)' in table 'user_info' doesn't exist", userInfo.GuildID)
	}

	delMemberInfo, retCode, err := DeleteGuildMember(guildInfo, targetMemberInfo)
	if err != nil {
		return retCode, err
	}

	var targetUserInfo table.TblUserInfo
	retCode, err = GetUserInfo(targetMemberInfo.MemberUin, &targetUserInfo)
	if err != nil {
		return retCode, err
	}

	var responseData proto.ProtoKickOutOfGuildResponse

	// 公会信息
	guildVitality, err := GetGuildVitality(userInfo.GuildID)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	var chairman table.TblUserInfo
	retCode, err = GetUserInfo(guildInfo.Chairman, &chairman)
	if err != nil {
		return retCode, err
	}

	responseData.GuildInfo = new(proto.ProtoGuildInfo)
	retCode, err = composeProtoGuildInfo(responseData.GuildInfo, guildInfo, &chairman, guildVitality)
	if err != nil {
		return retCode, err
	}

	// 被删除公会成员信息
	leagueInfo, err := GetLeagueInfo(this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	tblUserheadInfo, err := GetUserHeadFrame(this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	protoGuildMemberInfo := new(proto.ProtoGuildMemberInfo)
	composeProtoGuildMemberInfo(protoGuildMemberInfo, delMemberInfo, &targetUserInfo, leagueInfo, 0, tblUserheadInfo)
	responseData.GuildMembers = append(responseData.GuildMembers, protoGuildMemberInfo)

	// 发送聊天系统消息
	protoMessageInfo, err := AddMemberPostChangedMessage(userInfo.GuildID, this.Request.Uin, &userInfo, &targetUserInfo, "", MEMBER_POST_OPERATION_KICK)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	responseData.MessageList = append(responseData.MessageList, protoMessageInfo)

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *GuildHandler) PromoteMember() (int, error) {
	var reqParams proto.ProtoPromoteMemberRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if reqParams.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	if this.Request.Uin == reqParams.Uin {
		return errorcode.ERROR_CODE_GUILD_PERMISSION_DENIED, custom_errors.New("user can not modify his/her own post")
	}

	var userInfo table.TblUserInfo
	_, err = GetUserInfo(this.Request.Uin, &userInfo)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	creator, id, ok := ConvertGuildIdToUinAndId(userInfo.GuildID)
	if !ok {
		return errorcode.ERROR_CODE_GUILD_PERMISSION_DENIED, custom_errors.New("user is not in guild")
	}

	locker, err := LockGuild(creator, id)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	defer UnlockGuild(creator, id, locker)

	locUserGuildInfo, err := GetGuildMemberInfo(creator, id, this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if locUserGuildInfo == nil {
		return errorcode.ERROR_CODE_GUILD_PERMISSION_DENIED, custom_errors.New("user is not in current guild")
	}

	modifyUserGuildInfo, err := GetGuildMemberInfo(creator, id, reqParams.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if modifyUserGuildInfo == nil {
		return errorcode.ERROR_CODE_GUILD_NOT_IN_THIS_GUILD, custom_errors.New("modified user is not in current guild")
	}

	var responseData proto.ProtoPromoteMemberResponse

	if locUserGuildInfo.Post == GUILD_POST_CHAIRMAN && modifyUserGuildInfo.Post == GUILD_POST_VICE_CHAIRMAN {
		locUserGuildInfo.Post = modifyUserGuildInfo.Post
		modifyUserGuildInfo.Post = GUILD_POST_CHAIRMAN

		retCode, err := UpdateGuildMemberInfo(creator, id, locUserGuildInfo)
		if err != nil {
			return retCode, err
		}

		retCode, err = UpdateGuildMemberInfo(creator, id, modifyUserGuildInfo)
		if err != nil {
			return retCode, err
		}

		guildInfoModel := model.GuildInfoModel{Creator: creator, Id: id}
		guildInfo, err := guildInfoModel.GetGuildInfo()
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		guildInfo.Chairman = modifyUserGuildInfo.MemberUin
		retCode, err = guildInfoModel.UpdateGuildInfo(guildInfo)
		if err != nil {
			return retCode, err
		}

		vitality, err := GetGuildVitality(userInfo.GuildID)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		leagueInfo, err := GetLeagueInfo(this.Request.Uin)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		weeklyVitality, err := GetGuildWeeklyVitality(creator, id, this.Request.Uin)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		tblUserheadInfo, err := GetUserHeadFrame(userInfo.Uin)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		protoLocGuildMemberInfo := new(proto.ProtoGuildMemberInfo)
		composeProtoGuildMemberInfo(protoLocGuildMemberInfo, locUserGuildInfo, &userInfo, leagueInfo, weeklyVitality, tblUserheadInfo)
		responseData.GuildMembers = append(responseData.GuildMembers, protoLocGuildMemberInfo)

		var targetUserInfo table.TblUserInfo
		retCode, err = GetUserInfo(reqParams.Uin, &targetUserInfo)
		if err != nil {
			return retCode, err
		}
		retCode, err = composeProtoGuildInfo(&responseData.GuildInfo, guildInfo, &targetUserInfo, vitality)
		if err != nil {
			return retCode, err
		}

		targetUserLeagueInfo, err := GetLeagueInfo(reqParams.Uin)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		weeklyVitality, err = GetGuildWeeklyVitality(creator, id, this.Request.Uin)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		targetUserheadInfo, err := GetUserHeadFrame(targetUserInfo.Uin)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		protoModifiedGuildMemberInfo := new(proto.ProtoGuildMemberInfo)
		composeProtoGuildMemberInfo(protoModifiedGuildMemberInfo, modifyUserGuildInfo, &targetUserInfo, targetUserLeagueInfo, weeklyVitality, targetUserheadInfo)
		responseData.GuildMembers = append(responseData.GuildMembers, protoModifiedGuildMemberInfo)

		protoMessageInfo, err := AddMemberPostChangedMessage(userInfo.GuildID, this.Request.Uin, &userInfo, &targetUserInfo, modifyUserGuildInfo.Post, MEMBER_POST_OPERATION_PROMOTE)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
		responseData.MessageList = append(responseData.MessageList, protoMessageInfo)

		this.Response.ResData.Params = responseData
	} else {
		if locUserGuildInfo.Post != GUILD_POST_CHAIRMAN && (GuildPostSort[modifyUserGuildInfo.Post]-GuildPostSort[locUserGuildInfo.Post]) < 2 {
			return errorcode.ERROR_CODE_GUILD_PERMISSION_DENIED, custom_errors.New("user has no permission to promote %s's post", modifyUserGuildInfo.Post)
		}

		postSort := GuildPostSort[modifyUserGuildInfo.Post]
		targetPost := GuildPostSort2Post[postSort-1]

		membersByPost, err := GetGuildMemberByPost(creator, id, targetPost)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		targetPostLimit, ok := config.GGlobalConfig.Guild.LimitPostMemberCount[targetPost]

		if ok && targetPostLimit > 0 && len(membersByPost) >= targetPostLimit {
			return errorcode.ERROR_CODE_GUILD_POST_FULL, custom_errors.New("target post member is full")
		}

		modifyUserGuildInfo.Post = targetPost
		retCode, err := UpdateGuildMemberInfo(creator, id, modifyUserGuildInfo)
		if err != nil {
			return retCode, err
		}

		guildInfoModel := model.GuildInfoModel{Creator: creator, Id: id}
		guildInfo, err := guildInfoModel.GetGuildInfo()
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
		vitality, err := GetGuildVitality(userInfo.GuildID)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
		retCode, err = composeProtoGuildInfo(&responseData.GuildInfo, guildInfo, nil, vitality)
		if err != nil {
			return retCode, err
		}

		var targetUserInfo table.TblUserInfo
		retCode, err = GetUserInfo(reqParams.Uin, &targetUserInfo)
		if err != nil {
			return retCode, err
		}

		targetUserLeagueInfo, err := GetLeagueInfo(reqParams.Uin)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		weeklyVitality, err := GetGuildWeeklyVitality(creator, id, reqParams.Uin)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		targetUserheadInfo, err := GetUserHeadFrame(targetUserInfo.Uin)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		protoModifiedGuildMemberInfo := new(proto.ProtoGuildMemberInfo)
		composeProtoGuildMemberInfo(protoModifiedGuildMemberInfo, modifyUserGuildInfo, &targetUserInfo, targetUserLeagueInfo, weeklyVitality, targetUserheadInfo)
		responseData.GuildMembers = append(responseData.GuildMembers, protoModifiedGuildMemberInfo)

		protoMessageInfo, err := AddMemberPostChangedMessage(userInfo.GuildID, this.Request.Uin, &userInfo, &targetUserInfo, modifyUserGuildInfo.Post, MEMBER_POST_OPERATION_PROMOTE)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
		responseData.MessageList = append(responseData.MessageList, protoMessageInfo)

		this.Response.ResData.Params = responseData
	}

	return 0, nil
}

func (this *GuildHandler) DemoteMember() (int, error) {
	var reqParams proto.ProtoDemoteMemberRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if reqParams.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	if reqParams.Uin == this.Request.Uin {
		return errorcode.ERROR_CODE_GUILD_PERMISSION_DENIED, custom_errors.New("user can not modify his/her own post")
	}

	var userInfo table.TblUserInfo
	retCode, err := GetUserInfo(this.Request.Uin, &userInfo)
	if err != nil {
		return retCode, err
	}

	creator, gId, ok := ConvertGuildIdToUinAndId(userInfo.GuildID)
	if !ok {
		return errorcode.ERROR_CODE_GUILD_PERMISSION_DENIED, custom_errors.New("user is not in guild")
	}

	locker, err := LockGuild(creator, gId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	defer UnlockGuild(creator, gId, locker)

	targetMemberInfo, err := GetGuildMemberInfo(creator, gId, reqParams.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if targetMemberInfo == nil {
		return errorcode.ERROR_CODE_GUILD_NOT_IN_THIS_GUILD, custom_errors.New("target user is not in this guild")
	}

	memberInfo, err := GetGuildMemberInfo(creator, gId, this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	ps := GuildPostSort[memberInfo.Post]
	targetPS := GuildPostSort[targetMemberInfo.Post]
	targetFinalPost, ok := GuildPostSort2Post[targetPS+1]

	if ps >= targetPS || !ok {
		return errorcode.ERROR_CODE_GUILD_PERMISSION_DENIED, custom_errors.New("has no permission to demote %s", targetMemberInfo.Post)
	}

	targetMemberInfo.Post = targetFinalPost
	retCode, err = UpdateGuildMemberInfo(creator, gId, targetMemberInfo)
	if err != nil {
		return retCode, err
	}

	var targetUserInfo table.TblUserInfo
	retCode, err = GetUserInfo(reqParams.Uin, &targetUserInfo)
	if err != nil {
		return retCode, err
	}

	targetLeagueInfo, err := GetLeagueInfo(reqParams.Uin)
	if err != nil {
		return retCode, err
	}
	weeklyVitality, err := GetGuildWeeklyVitality(creator, gId, reqParams.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	tblUserheadInfo, err := GetUserHeadFrame(targetUserInfo.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	protoGuildMemberInfo := new(proto.ProtoGuildMemberInfo)
	composeProtoGuildMemberInfo(protoGuildMemberInfo, targetMemberInfo, &targetUserInfo, targetLeagueInfo, weeklyVitality, tblUserheadInfo)

	var responseData proto.ProtoDemoteMemberResponse
	responseData.GuildMembers = append(responseData.GuildMembers, protoGuildMemberInfo)

	guildInfoModel := model.GuildInfoModel{Creator: creator, Id: gId}
	guildInfo, err := guildInfoModel.GetGuildInfo()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	vitality, err := GetGuildVitality(userInfo.GuildID)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	composeProtoGuildInfo(&responseData.GuildInfo, guildInfo, nil, vitality)

	protoMessageInfo, err := AddMemberPostChangedMessage(userInfo.GuildID, this.Request.Uin, &userInfo, &targetUserInfo, targetMemberInfo.Post, MEMBER_POST_OPERATION_DEMOTE)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	responseData.MessageList = append(responseData.MessageList, protoMessageInfo)

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *GuildHandler) GuildMemberUinList() (int, error) {
	var reqParams proto.ProtoGetGuildMemberUinListRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	creator, gId, ok := ConvertGuildIdToUinAndId(reqParams.GuildId)
	if !ok {
		return errorcode.ERROR_CODE_GUILD_ID_FORMAT_WRONG, custom_errors.New("guild id format error")
	}

	allMembers, err := GetAllGuildMemberInfo(creator, gId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	var responseData proto.ProtoGetGuildMemeberUinListResponse
	responseData.UinList = make([]int, 0, len(allMembers))
	for _, member := range allMembers {
		responseData.UinList = append(responseData.UinList, member.MemberUin)
	}

	this.Response.ResData.Params = responseData

	return 0, nil
}

func CheckChairmanOfflineTime(guildInfo *table.TblGuildInfo, chairman *table.TblUserInfo) (*table.TblUserInfo, error) {
	if guildInfo == nil || chairman == nil {
		return nil, custom_errors.NullPoint()
	}

	bOnline, _ := IsUserOnline(chairman.Uin)
	base.GLog.Debug("IsUserOnline Uin %d result %t", chairman.Uin, bOnline)
	if bOnline {
		return nil, nil
	}

	offlineTime := int(base.GLocalizedTime.SecTimeStamp()) - chairman.UserOfflineTime
	base.GLog.Debug("chairman[%d] dboffline: %d offline time: %d max_offline_time: %d, guild[%d] memberCount[%d]", chairman.Uin, chairman.UserOfflineTime, offlineTime,
		config.GGlobalConfig.Guild.ChairmanMaxOfflineTime, guildInfo.Id, guildInfo.MemberCount)

	if offlineTime >= config.GGlobalConfig.Guild.ChairmanMaxOfflineTime && guildInfo.MemberCount > 1 {
		locker, err := LockGuild(guildInfo.Creator, guildInfo.Id)
		if err != nil {
			return nil, err
		}
		defer UnlockGuild(guildInfo.Creator, guildInfo.Id, locker)

		allMembers, err := GetAllGuildMemberInfo(guildInfo.Creator, guildInfo.Id)
		if err != nil {
			return nil, err
		}

		var chairmanMemberInfo *table.TblGuildMemberInfo
		memberMap := make(map[int]*table.TblGuildMemberInfo)
		uinList := make([]int, 0, len(allMembers))

		for _, member := range allMembers {
			if member.MemberUin != chairman.Uin {
				uinList = append(uinList, member.MemberUin)
				memberMap[member.MemberUin] = member
			} else {
				chairmanMemberInfo = member
			}
		}

		weeklyVitality, err := GetGuildWeeklyVitalityOfAll(guildInfo.Creator, guildInfo.Id)
		if err != nil {
			return nil, err
		}

		userList, err := GetMultiUserInfo(uinList...)
		if err != nil {
			return nil, err
		}

		var nextChairman *table.TblGuildMemberInfo
		var nextChairmanUserInfo *table.TblUserInfo
		for _, userInfo := range userList {
			base.GLog.Debug("query user info is[%v]", userInfo)

			memberOfflineTime := int(base.GLocalizedTime.SecTimeStamp()) - userInfo.UserOfflineTime
			if memberOfflineTime < config.GGlobalConfig.Guild.ChairmanMaxOfflineTime {
				if nextChairman == nil {
					nextChairman = memberMap[userInfo.Uin]
					nextChairmanUserInfo = userInfo
					continue
				}

				if memberInfo, ok := memberMap[userInfo.Uin]; ok {
					if GuildPostSort[memberInfo.Post] > GuildPostSort[nextChairman.Post] {
						continue
					}

					if GuildPostSort[memberInfo.Post] == GuildPostSort[nextChairman.Post] {
						if weeklyVitality.MemberVitality[memberInfo.MemberUin] < weeklyVitality.MemberVitality[nextChairman.MemberUin] {
							continue
						}

						if weeklyVitality.MemberVitality[memberInfo.MemberUin] == weeklyVitality.MemberVitality[nextChairman.MemberUin] {
							if memberInfo.Id > nextChairman.Id {
								continue
							}
						}
					}

					nextChairman = memberInfo
					nextChairmanUserInfo = userInfo
				}
			}
		}

		if nextChairman != nil {
			nextChairman.Post = GUILD_POST_CHAIRMAN
			chairmanMemberInfo.Post = GUILD_POST_MEMBER
			guildInfo.Chairman = nextChairman.MemberUin

			base.GLog.Error("Guild[%d] chairMan[%d] ===> newChairMain[%d]", guildInfo.Id, chairman.Uin, nextChairmanUserInfo.Uin)

			_, err := UpdateGuildInfo(guildInfo)
			if err != nil {
				return nil, err
			}

			_, err = UpdateGuildMemberInfo(guildInfo.Creator, guildInfo.Id, chairmanMemberInfo)
			if err != nil {
				return nil, err
			}

			_, err = UpdateGuildMemberInfo(guildInfo.Creator, guildInfo.Id, nextChairman)
			if err != nil {
				return nil, err
			}
		} else {
			base.GLog.Error("Guild[%d] nextChairman is null", guildInfo.Id)
		}

		return nextChairmanUserInfo, nil
	}

	return nil, nil
}

// 验证
func ValidUserCreateGuildPermissionAndCost(userInfo *table.TblUserInfo) ([]config.ResourceItem, int, error) {
	if userInfo == nil {
		return nil, errorcode.ERROR_CODE_DEFAULT, custom_errors.New("user info is empty")
	}

	// 判断玩家是否存在公会
	if ValidGuildId(userInfo.GuildID) {
		return nil, errorcode.ERROR_CODE_GUILD_ALREADY_IN_GUILD, custom_errors.New("player[%d] is already in guild[%s]", userInfo.Uin, userInfo.GuildID)
	}

	// 判断玩家是否已解锁公会功能
	retCode, err := ValidUserGuildUnlocked(userInfo)
	if err != nil {
		return nil, retCode, err
	}

	// 判断资源是否足够
	realCost, retCode, err := CalculateUserRealResourcesCost(userInfo, &config.GGlobalConfig.Guild.CreateGuildCost.ResourceItems)
	if err != nil {
		return nil, retCode, err
	}

	return realCost, 0, nil
}

func ValidUserJoinGuildPermission(userInfo *table.TblUserInfo, creatorUin int, id int) (int, error) {
	if userInfo == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("user info is empty")
	}

	// 判断玩家是否已解锁公会功能
	retCode, err := ValidUserGuildUnlocked(userInfo)
	if err != nil {
		return retCode, err
	}

	// 判断玩家是否已经在公会中
	currentCreator, currentId, isValid := ConvertGuildIdToUinAndId(userInfo.GuildID)
	if isValid {
		if currentCreator == creatorUin && currentId == id {
			return errorcode.ERROR_CODE_GUILD_ALREADY_IN_THIS_GUILD, custom_errors.New("player[%d] is already in this guild[%s]", userInfo.Uin, userInfo.GuildID)
		} else {
			return errorcode.ERROR_CODE_GUILD_ALREADY_IN_GUILD, custom_errors.New("player[%d] is already in other guild[%s]", userInfo.Uin, userInfo.GuildID)
		}
	}

	return 0, nil

}

func ValidGuildJoinStatus(guildId string, guildInfo *table.TblGuildInfo, isInvited bool) (int, error) {
	if guildInfo == nil {
		return errorcode.ERROR_CODE_GUILD_ID_NOT_EXIST, custom_errors.New("guild[%s] is not exist", guildId)
	}

	if guildInfo.MemberCount >= config.GGlobalConfig.Guild.MaxMemberCount {
		return errorcode.ERROR_CODE_GUILD_MEMBER_FULL, custom_errors.New("guild[%s] member is full", guildId)
	}

	return 0, nil
}

func ValidUserGuildUnlocked(userInfo *table.TblUserInfo) (int, error) {
	if userInfo == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("userinfo is empty")
	}

	// 判断玩家是否已解锁公会功能
	if levelExpConf, ok := config.GLevelExpConfig.UnlockMap[config.LEVEL_EXP_UNLOCK_GUILD]; ok {
		if userInfo.Level < levelExpConf.Level {
			return errorcode.ERROR_CODE_GUILD_PLAYER_LEVEL_NOT_REACH, custom_errors.New("player level(%d) not reach %d", userInfo.Level, levelExpConf.Level)
		}
	} else {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("LevelExp.json does not contain unlock guild")
	}

	return 0, nil
}

func ValidCreateGuildInfo(guildInfo *proto.ProtoCreateGuildRequest) (int, error) {
	// 名字长度限制
	guildNameLen := len(guildInfo.GuildInfo.Name)
	if guildNameLen <= 0 || guildNameLen > 128 {
		return errorcode.ERROR_CODE_GUILD_NAME_LENGTH_INVALID, custom_errors.New("guild name length is invalid")
	}

	// 可修改项合法性检测
	retCode, err := ValidGuildModifiableInfo(&guildInfo.GuildInfo.ProtoGuildModifiableInfo)
	if err != nil {
		return retCode, err
	}

	return 0, nil
}

func ValidGuildModifiableInfo(info *proto.ProtoGuildModifiableInfo) (int, error) {
	base.GLog.Debug("GuildInfo: [%+v]", info)
	if info == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	// 加入类型判断
	if info.JoinType != GUILD_JOIN_TYPE_ALLOW_ANYONE && info.JoinType != GUILD_JOIN_TYPE_NEED_AUDIT && info.JoinType != GUILD_JOIN_TYPE_NOT_ALLOWED {
		return errorcode.ERROR_CODE_GUILD_JOIN_TYPE_ERROR, custom_errors.New("guild join type is not one of ALLOW_ANYONE(0), NEED_AUDIT(1) or NOT_ALLOWED(2)")
	}

	// 描述长度限制
	if utf8.RuneCountInString(info.Description) > 100 {
		return errorcode.ERROR_CODE_GUILD_DESC_TOO_LONG, custom_errors.New("guild description is too long")
	}

	// 公会标识字符串过长
	if len(info.Symbol) > 512 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("guild symbol string is too long.")
	}

	return 0, nil
}

// 查询数据库中公会信息
func GetGuildInfo(creator, gId int) (*table.TblGuildInfo, error) {
	guildInfoModel := model.GuildInfoModel{Creator: creator, Id: gId}
	return guildInfoModel.GetGuildInfo()
}

// 更新数据库中的公会信息。
// 如果公会不存在(Id <= 0)，则插入一条新数据；
// 如果公会存在(Id > 0)，且公会成员数小于等于0，则删除公会；
// 如果公会存在(Id > 0)，公会成员数大于0，则更新公会信息；
func UpdateGuildInfo(guildInfo *table.TblGuildInfo) (int, error) {
	guildInfoModel := model.GuildInfoModel{Creator: guildInfo.Creator, Id: guildInfo.Id}
	if guildInfo.Id <= 0 {
		// 公会不存在
		retCode, err := guildInfoModel.AddNewGuild(guildInfo)
		if err != nil {
			return retCode, err
		}
	} else {
		if guildInfo.MemberCount > 0 {
			// 公会不存在且更新后成员数大于0
			retCode, err := guildInfoModel.UpdateGuildInfo(guildInfo)
			if err != nil {
				return retCode, err
			}
		} else {
			// 公会存在，但是成员数小于0，此时应该删除公会
			retCode, err := DeleteGuildInfo(guildInfo)
			if err != nil {
				return retCode, err
			}
		}
	}

	return 0, nil
}

// 删除数据库中公会信息
// 同时删除公会申请、公会活跃度、公会护卫舰、公会聊天记录
func DeleteGuildInfo(guildInfo *table.TblGuildInfo) (int, error) {
	if guildInfo == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("guild info is nil")
	}

	guildId := FormatGuildId(guildInfo.Creator, guildInfo.Id)

	// 1. 删除公会信息
	guildInfoModel := model.GuildInfoModel{Creator: guildInfo.Creator, Id: guildInfo.Id}
	retCode, err := guildInfoModel.DeleteGuildInfo(guildInfo)
	if err != nil {
		return retCode, err
	}

	// 2. 删除公会申请信息
	retCode, err = DeleteAllApplyInfo(guildInfo.Creator, guildInfo.Id)
	if err != nil {
		return retCode, err
	}

	// 3. 删除公会活跃度
	retCode, err = DeleteGuildVitality(guildId)
	if err != nil {
		return retCode, err
	}

	// 4. 删除公会护卫舰
	retCode, err = DeleteGuildFrigateShip(guildInfo.Id, guildInfo.Creator)
	if err != nil {
		return retCode, err
	}

	// 5. 删除公会聊天记录
	retCode, err = DeleteChannel(guildId)
	if err != nil {
		return retCode, err
	}

	return 0, nil
}

// 设置可修改信息
func SetGuildModifiableInfo(target *table.TblGuildInfo, modifyInfo *proto.ProtoGuildModifiableInfo) {
	if target == nil {
		return
	}

	target.Symbol = modifyInfo.Symbol
	target.JoinType = modifyInfo.JoinType
	target.Description = modifyInfo.Description
	target.CondLeagueLevel = modifyInfo.CondLeagueLevel
}

func AddMemeberCountChangedMessage(channel string, uin int, userInfo *table.TblUserInfo, operation int) (*proto.ProtoMessageInfo, error) {
	protoMessageContent := new(proto.ProtoGuildMemberCountChangedMessage)
	protoMessageContent.UserInfo.UserName = userInfo.UserName
	protoMessageContent.UserInfo.Icon = userInfo.Icon
	protoMessageContent.UserInfo.Level = userInfo.Level
	protoMessageContent.Operation = operation

	data, err := json.Marshal(protoMessageContent)
	if err != nil {
		return nil, err
	}

	return AddMessage(channel, uin, string(data), CHAT_MESSAGE_TYPE_MEMBER_COUNT_CHANGED)
}

func AddMemberPostChangedMessage(channel string, uin int, operator *table.TblUserInfo, target *table.TblUserInfo, finalPost string, operation int) (*proto.ProtoMessageInfo, error) {
	protoMessageContent := new(proto.ProtoGuildMemberPostChangedMessage)
	protoMessageContent.FinalPost = finalPost
	protoMessageContent.Operation = operation

	protoMessageContent.Operator.UserName = operator.UserName
	protoMessageContent.Operator.Icon = operator.Icon
	protoMessageContent.Operator.Level = operator.Level

	protoMessageContent.TargetUser.UserName = target.UserName
	protoMessageContent.TargetUser.Icon = target.Icon
	protoMessageContent.TargetUser.Level = target.Level

	data, err := json.Marshal(protoMessageContent)
	if err != nil {
		return nil, err
	}

	return AddMessage(channel, uin, string(data), CHAT_MESSAGE_TYPE_MEMBER_POST_CHANGED)
}

func AddGuildChairmanTrasferMessage(channel string, uin int, oldChairman *table.TblUserInfo, newChairman *table.TblUserInfo) (*proto.ProtoMessageInfo, error) {
	protoMessageContent := new(proto.ProtoGuildChairmanTransferMessage)

	protoMessageContent.OldChairman.Uin = oldChairman.Uin
	protoMessageContent.OldChairman.UserName = oldChairman.UserName
	protoMessageContent.OldChairman.Icon = oldChairman.Icon
	protoMessageContent.OldChairman.Level = oldChairman.Level

	protoMessageContent.NewChairman.Uin = newChairman.Uin
	protoMessageContent.NewChairman.UserName = newChairman.UserName
	protoMessageContent.NewChairman.Icon = newChairman.Icon
	protoMessageContent.NewChairman.Level = newChairman.Level

	data, err := json.Marshal(protoMessageContent)
	if err != nil {
		return nil, err
	}

	return AddMessage(channel, uin, string(data), CHAT_MESSAGE_TYPE_CHAIRMAN_TRANSFER)
}

func NewGuild(uin int, guildName string) *table.TblGuildInfo {
	tblGuild := new(table.TblGuildInfo)
	tblGuild.Symbol = ""
	tblGuild.Description = ""
	tblGuild.JoinType = 0
	tblGuild.CondLeagueLevel = 0
	tblGuild.Name = guildName
	tblGuild.Level = 1
	tblGuild.Vitality = 0
	tblGuild.MemberCount = 0
	tblGuild.Creator = uin
	tblGuild.Chairman = uin
	tblGuild.CreateTime = int(base.GLocalizedTime.SecTimeStamp())

	return tblGuild
}

// 公会锁：同一时间仅允许有一个玩家操作公会信息
func LockGuild(creator int, gId int) (string, error) {
	k := fmt.Sprintf("%s.%d.%d", GUILD_LOCKER, creator, gId)
	return redistool.SpinLockWithFingerPoint(k, 0)
}

func UnlockGuild(creator int, gId int, value string) {
	k := fmt.Sprintf("%s.%d.%d", GUILD_LOCKER, creator, gId)
	err := redistool.UnLock(k, value)
	if err != nil {
		base.GLog.Error("unlock guild(creatorUin:%d, gId:%d, value:%s) error[%v]", creator, gId, value, err)
	}
}

// 用户公会信息锁：同一时间，某个玩家的GuildId只能被一个人修改
func LockUserGuild(uin int) (string, error) {
	k := fmt.Sprintf("%s.%d", GUILD_USER_LOCKER, uin)
	return redistool.SpinLockWithFingerPoint(k, 0)
}

func UnLockerUserGuild(uin int, value string) {
	k := fmt.Sprintf("%s.%d", GUILD_USER_LOCKER, uin)
	err := redistool.UnLock(k, value)
	if err != nil {
		base.GLog.Error("unlock user guild(uin:%d) failed. error[%v]", uin, err)
	}
}

// 公会信息协议
func composeProtoGuildInfo(target *proto.ProtoGuildInfo, data *table.TblGuildInfo, chairman *table.TblUserInfo, guildVitality int) (int, error) {
	if target == nil || data == nil {
		return 0, nil
	}

	target.GuildId = FormatGuildId(data.Creator, data.Id)
	target.Chairman = data.Chairman
	target.JoinType = data.JoinType
	target.CreateTime = data.CreateTime
	target.MemberCount = data.MemberCount
	target.CondLeagueLevel = data.CondLeagueLevel
	target.Name = data.Name
	target.Symbol = data.Symbol
	target.Description = data.Description
	target.PerformId = data.PerformId

	if chairman != nil {
		target.ChairmanName = chairman.UserName
	} else {
		chairman = new(table.TblUserInfo)
		retCode, err := GetUserInfo(data.Chairman, chairman)
		if err != nil {
			return retCode, err
		}
		target.ChairmanName = chairman.UserName
	}

	if guildVitality < 0 {
		vitality, err := GetGuildVitality(target.GuildId)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
		target.Vitality = vitality
	} else {
		target.Vitality = guildVitality
	}

	rank, err := GetGuildRank(target.GuildId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	target.Rank = rank + 1

	target.Level = config.GGuildLevelConfig.LevelBySumExp(target.Vitality)
	return 0, nil
}

func composeProtoGuildInfoWithNoCompletion(target *proto.ProtoGuildInfo, data *table.TblGuildInfo) (int, error) {
	if target == nil || data == nil {
		return 0, nil
	}

	target.GuildId = FormatGuildId(data.Creator, data.Id)
	target.Chairman = data.Chairman
	target.JoinType = data.JoinType
	target.CreateTime = data.CreateTime
	target.MemberCount = data.MemberCount
	target.CondLeagueLevel = data.CondLeagueLevel
	target.Name = data.Name
	target.Symbol = data.Symbol
	target.Description = data.Description
	target.PerformId = data.PerformId

	return 0, nil
}
