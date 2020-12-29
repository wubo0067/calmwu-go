package handler

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/dataaccess/redistool"
	"sailcraft/fleetsvr_main/config"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/model"
	"sailcraft/fleetsvr_main/proto"
	"sailcraft/fleetsvr_main/table"
)

const (
	GUILD_INVITE_LOCKER = "guild.invite.locker"
)

func (this *GuildHandler) SendInvite() (int, error) {
	var reqParams proto.ProtoGuildSendInviteRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if reqParams.TargetUin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("target uin is invalid")
	}

	// 判断发出邀请的玩家是否在公会中
	var userInfo table.TblUserInfo
	_, err = GetUserInfo(this.Request.Uin, &userInfo)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	creator, id, ok := ConvertGuildIdToUinAndId(userInfo.GuildID)
	if !ok {
		return errorcode.ERROR_CODE_GUILD_PERMISSION_DENIED, custom_errors.New("user is not in guild")
	}

	// 判断发出邀请的玩家是否有邀请其他玩家加入公会的权限
	guildMemberInfo, err := GetGuildMemberInfo(creator, id, this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if guildMemberInfo.Post != GUILD_POST_CHAIRMAN && guildMemberInfo.Post != GUILD_POST_VICE_CHAIRMAN {
		return errorcode.ERROR_CODE_GUILD_PERMISSION_DENIED, custom_errors.New("user has no permission to invite other player")
	}

	// 判断被邀请的玩家是否已经在公会中
	var targetUserInfo table.TblUserInfo
	_, err = GetUserInfo(reqParams.TargetUin, &targetUserInfo)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	targetCreator, targetGId, ok := ConvertGuildIdToUinAndId(targetUserInfo.GuildID)
	if ok {
		if targetCreator == creator && targetGId == id {
			return errorcode.ERROR_CODE_GUILD_ALREADY_IN_THIS_GUILD, custom_errors.New("target user is already in this guild")
		} else {
			return errorcode.ERROR_CODE_GUILD_ALREADY_IN_GUILD, custom_errors.New("target user is already in other guild")
		}
	}

	// 被邀请的玩家是否已经解锁公会
	retCode, err := ValidUserGuildUnlocked(&targetUserInfo)
	if err != nil {
		return retCode, err
	}

	locker, err := LockGuildInvite(reqParams.TargetUin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	defer UnlockGuildInvite(reqParams.TargetUin, locker)

	retCode, err = AddGuildInvite(reqParams.TargetUin, this.Request.Uin, creator, id)
	if err != nil {
		return retCode, err
	}

	return 0, nil
}

func (this *GuildHandler) GetInviteList() (int, error) {
	inviteList, err := GetAllGuildInvite(this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	var responseData proto.ProtoGuildGetAllInviteResponse
	responseData.InviteList = make([]*proto.ProtoGuildInviteInfo, 0, len(inviteList))
	if len(inviteList) <= 0 {
		this.Response.ResData.Params = responseData
		return 0, nil
	}

	guildIds := make([]string, 0, len(inviteList))
	uinSlice := make([]int, 0, len(inviteList))
	for _, invite := range inviteList {
		guildIds = append(guildIds, invite.GuildId)
		uinSlice = append(uinSlice, invite.FromUin)
	}

	guildInfoList, err := model.GetMultiGuildInfo(guildIds...)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	guildInfoMap := make(map[string]*table.TblGuildInfo)
	for _, guildInfo := range guildInfoList {
		guildId := FormatGuildId(guildInfo.Creator, guildInfo.Id)
		guildInfoMap[guildId] = guildInfo
	}

	userInfoList, err := model.GetMultiUserInfo(uinSlice...)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	userInfoMap := make(map[int]*table.TblUserInfo)
	for _, userInfo := range userInfoList {
		userInfoMap[userInfo.Uin] = userInfo
	}

	invalidInvites := make([]string, 0)
	for _, invite := range inviteList {
		guildInfo, ok := guildInfoMap[invite.GuildId]
		if !ok {
			invalidInvites = append(invalidInvites, invite.GuildId)
			continue
		}

		userInfo, ok := userInfoMap[invite.FromUin]
		if !ok {
			invalidInvites = append(invalidInvites, invite.GuildId)
			continue
		}

		protoInviteInfo := new(proto.ProtoGuildInviteInfo)
		retCode, err := composeProtoGuildInviteInfo(protoInviteInfo, invite, guildInfo, userInfo)
		if err != nil {
			return retCode, err
		}

		responseData.InviteList = append(responseData.InviteList, protoInviteInfo)
	}

	if len(invalidInvites) > 0 {
		DelGuildInvite(this.Request.Uin, invalidInvites...)
	}

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *GuildHandler) HandleInvite() (int, error) {
	var reqParams proto.ProtoHandleGuildInviteRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	creator, id, ok := ConvertGuildIdToUinAndId(reqParams.GuildId)
	if !ok {
		return errorcode.ERROR_CODE_GUILD_ID_FORMAT_WRONG, custom_errors.New("guild id format error")
	}

	locker, err := LockGuildInvite(this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	defer UnlockGuildInvite(this.Request.Uin, locker)

	inviteInfo, err := GetGuildInviteInfo(this.Request.Uin, reqParams.GuildId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if inviteInfo == nil {
		return errorcode.ERROR_CODE_GUILD_INVITE_NOT_EXIST, custom_errors.New("guild invite not exists")
	}

	if reqParams.Operation == GUILD_INVITE_OPERATION_REFUSE {
		retCode, err := DelGuildInvite(this.Request.Uin, reqParams.GuildId)
		if err != nil {
			return retCode, err
		}

		var responseData proto.ProtoHandleGuildInviteResponse
		responseData.Operation = reqParams.Operation
		responseData.MessageList = make([]*proto.ProtoMessageInfo, 0)

		this.Response.ResData.Params = responseData
	} else {
		var userInfo table.TblUserInfo
		_, err := GetUserInfo(this.Request.Uin, &userInfo)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		if ValidGuildId(userInfo.GuildID) {
			return errorcode.ERROR_CODE_GUILD_ALREADY_IN_GUILD, custom_errors.New("user is already in guild")
		}

		leaveInfo, err := GetGuildLeaveInfo(this.Request.Uin)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		if leaveInfo != nil {
			leftTime := config.GGlobalConfig.Guild.JoinGuildColdTime + leaveInfo.LeaveTime - int(base.GLocalizedTime.SecTimeStamp())
			if leftTime > 0 {
				// 这里虽然返回了错误码，也得把剩余时间返回回去
				var responseData proto.ProtoJoinGuildWaitForLeaveTimeResponse
				responseData.RestWaitTime = leftTime
				this.Response.ResData.Params = responseData

				return errorcode.ERROR_CODE_GUILD_JOIN_NEED_WAIT, custom_errors.New("cold time not reach")
			}
		}

		locker, err := LockGuild(creator, id)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
		defer UnlockGuild(creator, id, locker)

		guildInfoModel := model.GuildInfoModel{Creator: creator, Id: id}
		guildInfo, err := guildInfoModel.GetGuildInfo()
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		if guildInfo == nil {
			retCode, err := DelGuildInvite(this.Request.Uin, reqParams.GuildId)
			if err != nil {
				return retCode, err
			}

			return errorcode.ERROR_CODE_GUILD_ID_NOT_EXIST, custom_errors.New("guild not exist, may be dissolved")
		}

		if guildInfo.MemberCount >= config.GGlobalConfig.Guild.MaxMemberCount {
			return errorcode.ERROR_CODE_GUILD_MEMBER_FULL, custom_errors.New("guild member is full.")
		}

		_, _, err = AddNewGuildMember(guildInfo, &userInfo, GUILD_POST_MEMBER)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		_, err = DelAllGuildInvite(this.Request.Uin)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		vitality, err := GetGuildVitality(reqParams.GuildId)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		var responseData proto.ProtoHandleGuildInviteResponse
		responseData.Operation = reqParams.Operation
		responseData.GuildInfo = new(proto.ProtoGuildInfo)
		retCode, err := composeProtoGuildInfo(responseData.GuildInfo, guildInfo, nil, vitality)
		if err != nil {
			return retCode, err
		}

		guildMemberInfoModel := model.GuildMemberInfoModel{Creator: creator, GId: id}
		allMembers, err := guildMemberInfoModel.GetAllGuildMember()
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		uinList := make([]int, 0, len(allMembers))
		for _, member := range allMembers {
			uinList = append(uinList, member.MemberUin)
		}

		allUsers, err := GetMultiUserInfo(uinList...)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		allLeagues, err := GetMultiLeagueInfo(uinList...)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		userMap := make(map[int]*table.TblUserInfo)
		leagueMap := make(map[int]*table.TblLeagueInfo)

		for _, user := range allUsers {
			userMap[user.Uin] = user
		}

		for _, league := range allLeagues {
			leagueMap[league.Uin] = league
		}

		guildWeeklyVitality, err := GetGuildWeeklyVitalityOfAll(creator, id)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		userHeadFrameMap, err := GetMultiUserHeadFrame(uinList...)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		for _, member := range allMembers {
			memWeeklyVitality, ok := guildWeeklyVitality.MemberVitality[member.MemberUin]
			if !ok {
				memWeeklyVitality = 0
			}

			protoMemberInfo := new(proto.ProtoGuildMemberInfo)
			retCode, err := composeProtoGuildMemberInfo(protoMemberInfo, member, userMap[member.MemberUin], leagueMap[member.MemberUin], memWeeklyVitality, userHeadFrameMap[member.MemberUin])
			if err != nil {
				base.GLog.Debug("ComposeProtoGuildMemberInfo Error: \nUserInfo:%+v\nLeagueInfo:%+v\n", userMap[member.MemberUin], leagueMap[member.MemberUin])
				return retCode, err
			}

			responseData.GuildMembers = append(responseData.GuildMembers, protoMemberInfo)
		}

		protoMessageInfo, err := AddMemeberCountChangedMessage(reqParams.GuildId, this.Request.Uin, &userInfo, MEMBER_COUNT_OPERATION_JOIN)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
		responseData.MessageList = append(responseData.MessageList, protoMessageInfo)

		this.Response.ResData.Params = responseData
	}

	return 0, nil
}

func AddGuildInvite(targetUin, fromUin, creator, gId int) (int, error) {
	if targetUin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("target uin is invalid")
	}

	if fromUin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("from uin is invalid")
	}

	if creator <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("creator is invalid")
	}

	if gId <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("gId is invalid")
	}

	guildId := FormatGuildId(creator, gId)

	guildInviteModel := model.GuildInviteInfoModel{Uin: targetUin}

	// 判断邀请信息是否存在
	inviteInfo, err := guildInviteModel.GetGuildInviteInfoByGuildId(guildId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if inviteInfo != nil {
		return errorcode.ERROR_CODE_GUILD_INVITE_REPEAT, custom_errors.New("send guild invite repeat")
	}

	// 添加邀请信息
	inviteInfo = new(table.TblGuildInvite)
	inviteInfo.FromUin = fromUin
	inviteInfo.InviteTime = int(base.GLocalizedTime.SecTimeStamp())
	inviteInfo.Uin = targetUin
	inviteInfo.GuildId = FormatGuildId(creator, gId)
	inviteCount, err := guildInviteModel.AddGuildInvite(inviteInfo)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	// 删除多余元素
	delCount := inviteCount - config.GGlobalConfig.Guild.LimitGuildInvite
	if delCount > 0 {
		retCode, err := guildInviteModel.DeleteOldestInvitInfo(delCount)
		if err != nil {
			return retCode, err
		}
	}

	return 0, nil
}

func DelGuildInvite(targetUin int, guildId ...string) (int, error) {
	guildInviteModel := model.GuildInviteInfoModel{Uin: targetUin}
	return guildInviteModel.DeleteGuildInviteByGuildId(guildId...)
}

func DelAllGuildInvite(targetUin int) (int, error) {
	guildInviteModel := model.GuildInviteInfoModel{Uin: targetUin}
	return guildInviteModel.DeleteAllInvite()
}

func GetAllGuildInvite(targetUin int) ([]*table.TblGuildInvite, error) {
	guildInviteModel := model.GuildInviteInfoModel{Uin: targetUin}
	return guildInviteModel.GetAllGuildInvite()
}

func GetGuildInviteInfo(targetUin int, guildId string) (*table.TblGuildInvite, error) {
	guildInviteModel := model.GuildInviteInfoModel{Uin: targetUin}
	return guildInviteModel.GetGuildInviteInfoByGuildId(guildId)
}

func LockGuildInvite(uin int) (string, error) {
	key := fmt.Sprintf("%s.%d", GUILD_INVITE_LOCKER, uin)
	return redistool.SpinLockWithFingerPoint(key, 0)
}

func UnlockGuildInvite(uin int, value string) {
	key := fmt.Sprintf("%s.%d", GUILD_INVITE_LOCKER, uin)
	err := redistool.UnLock(key, value)
	if err != nil {
		base.GLog.Error("unlock %s error[%s]", key, err)
	}
}

func composeProtoGuildInviteInfo(target *proto.ProtoGuildInviteInfo, inviteInfo *table.TblGuildInvite, guildInfo *table.TblGuildInfo, userInfo *table.TblUserInfo) (int, error) {
	if target == nil || guildInfo == nil || userInfo == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	target.FromUin = inviteInfo.FromUin
	target.FromUserName = userInfo.UserName
	target.FromIcon = userInfo.Icon
	target.FromCountryCode = userInfo.ISOCountryCode
	target.Invitetime = inviteInfo.InviteTime
	target.TargetUin = inviteInfo.Uin
	target.GuildInfo = new(proto.ProtoGuildInfo)

	vitality, err := GetGuildVitality(FormatGuildId(guildInfo.Creator, guildInfo.Id))
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	retCode, err := composeProtoGuildInfo(target.GuildInfo, guildInfo, nil, vitality)
	if err != nil {
		return retCode, err
	}

	return 0, nil
}
