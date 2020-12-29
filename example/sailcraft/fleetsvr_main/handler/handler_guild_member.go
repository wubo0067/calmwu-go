package handler

import (
	"sailcraft/base"
	"sailcraft/fleetsvr_main/config"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/model"
	"sailcraft/fleetsvr_main/proto"
	"sailcraft/fleetsvr_main/table"
)

func (this *GuildHandler) GetGuildMemberInfo() (int, error) {
	var reqParams proto.ProtoGetGuildMemberInfoByUinRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if reqParams.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.InvalidUin()
	}

	creator, gId, ok := ConvertGuildIdToUinAndId(reqParams.GuildId)
	if !ok {
		return errorcode.ERROR_CODE_GUILD_ID_FORMAT_WRONG, custom_errors.New("guild id [%s] format error", reqParams.GuildId)
	}

	guildMemberInfo, err := GetGuildMemberInfo(creator, gId, reqParams.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	tblUserheadInfo, err := GetUserHeadFrame(reqParams.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	weeklyVitality, err := GetGuildWeeklyVitality(creator, gId, reqParams.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	protoGuildMemberInfo := new(proto.ProtoGuildMemberInfo)
	composeProtoGuildMemberInfo(protoGuildMemberInfo, guildMemberInfo, nil, nil, weeklyVitality, tblUserheadInfo)

	var responseData proto.ProtoGetGuildMemberInfoByUinResponse
	responseData.GuildMemberInfo = protoGuildMemberInfo

	this.Response.ResData.Params = responseData

	return 0, nil
}

func AddNewGuildMember(guildInfo *table.TblGuildInfo, userInfo *table.TblUserInfo, post string) (*table.TblGuildMemberInfo, int, error) {
	if guildInfo == nil || userInfo == nil {
		return nil, errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	guildInfo.MemberCount += 1
	retCode, err := UpdateGuildInfo(guildInfo)
	if err != nil {
		return nil, retCode, err
	}

	memberInfo, err := AddNewGuildMemberInfo(guildInfo.Creator, guildInfo.Id, userInfo.Uin, post)
	if err != nil {
		return nil, errorcode.ERROR_CODE_DEFAULT, err
	}

	return memberInfo, 0, nil
}

func DeleteGuildMemberByUin(guildInfo *table.TblGuildInfo, memberUin int) (*table.TblGuildMemberInfo, int, error) {
	if guildInfo == nil {
		return nil, errorcode.ERROR_CODE_GUILD_ID_NOT_EXIST, custom_errors.New("guild not exist")
	}

	delMemInfo, err := GetGuildMemberInfo(guildInfo.Creator, guildInfo.Id, memberUin)
	if err != nil {
		return nil, errorcode.ERROR_CODE_DEFAULT, err
	}

	return DeleteGuildMember(guildInfo, delMemInfo)
}

func DeleteGuildMember(guildInfo *table.TblGuildInfo, delMemInfo *table.TblGuildMemberInfo) (*table.TblGuildMemberInfo, int, error) {
	if guildInfo == nil {
		return nil, errorcode.ERROR_CODE_GUILD_ID_NOT_EXIST, custom_errors.New("guild not exist")
	}

	if delMemInfo == nil {
		return nil, errorcode.ERROR_CODE_GUILD_NOT_IN_THIS_GUILD, custom_errors.New("user is not in this guild")
	}

	// Prepare： 公会战期间不允许将玩家踢出公会
	/*
		guildWardInfo, err := GetGuildWar()
		if err != nil {
			return nil, errorcode.ERROR_CODE_DEFAULT, err
		}

		if guildWardInfo.Phase != GUILD_WAR_PHASE_STOP {
			return nil, errorcode.ERROR_CODE_GUILD_IN_WAR, custom_errors.New("can not delete member during guild ward")
		}
	*/

	guildMemberInfoModel := model.GuildMemberInfoModel{Creator: guildInfo.Creator, GId: guildInfo.Id}
	memberUin := delMemInfo.MemberUin
	// 1. 删除成员信息, 如果删除的成员是会长，则给剩下的成员先按职位排序，然后按排名排序，选择最高名次的成员为会长，更新该成员信息
	retCode, err := guildMemberInfoModel.DeleteMember(delMemInfo)
	if err != nil {
		return nil, retCode, err
	}

	// 2. 删除玩家周活跃度
	retCode, err = DeleteGuildWeeklyVitalityOfMember(guildInfo.Creator, guildInfo.Id, delMemInfo.MemberUin)
	if err != nil {
		return nil, errorcode.ERROR_CODE_DEFAULT, err
	}

	if delMemInfo.Post == GUILD_POST_CHAIRMAN && guildInfo.MemberCount > 1 {
		allMembers, err := guildMemberInfoModel.GetAllGuildMember()
		if err != nil {
			return nil, errorcode.ERROR_CODE_DEFAULT, err
		}

		weeklyVitality, err := GetGuildWeeklyVitalityOfAll(guildInfo.Creator, guildInfo.Id)
		if err != nil {
			return delMemInfo, errorcode.ERROR_CODE_DEFAULT, err
		}

		nextChairman := allMembers[0]
		for _, member := range allMembers[1:] {
			if GuildPostSort[member.Post] > GuildPostSort[nextChairman.Post] {
				continue
			}

			if GuildPostSort[member.Post] == GuildPostSort[nextChairman.Post] {
				if weeklyVitality.MemberVitality[member.MemberUin] < weeklyVitality.MemberVitality[nextChairman.MemberUin] {
					continue
				}

				if weeklyVitality.MemberVitality[member.MemberUin] == weeklyVitality.MemberVitality[nextChairman.MemberUin] {
					if member.Id > nextChairman.Id {
						continue
					}
				}
			}

			nextChairman = member
		}

		nextChairman.Post = GUILD_POST_CHAIRMAN
		retCode, err := guildMemberInfoModel.UpdateGuildMemberInfo(nextChairman)
		if err != nil {
			return nil, retCode, err
		}

		guildInfo.Chairman = nextChairman.MemberUin
	}

	// 3. 更新公会信息
	guildInfo.MemberCount -= 1
	retCode, err = UpdateGuildInfo(guildInfo)
	if err != nil {
		return nil, retCode, err
	}

	// 4. 更新用户离开公会时间
	record := table.TblGuildLeaveInfo{Uin: memberUin, GuildId: FormatGuildId(guildInfo.Creator, guildInfo.Id), LeaveTime: int(base.GLocalizedTime.SecTimeStamp())}
	retCode, err = SetGuildLeaveInfo(memberUin, &record)
	if err != nil {
		return nil, retCode, err
	}

	return delMemInfo, 0, nil
}

func GetGuildMemberInfo(creator, gId, uin int) (*table.TblGuildMemberInfo, error) {
	guildMemberInfoModel := model.GuildMemberInfoModel{Creator: creator, GId: gId}
	return guildMemberInfoModel.GetGuildMemberInfo(uin)
}

func GetAllGuildMemberInfo(creator, gId int) ([]*table.TblGuildMemberInfo, error) {
	guildMemberInfoModel := model.GuildMemberInfoModel{Creator: creator, GId: gId}
	return guildMemberInfoModel.GetAllGuildMember()
}

func AddNewGuildMemberInfo(creator, gId, uin int, post string) (*table.TblGuildMemberInfo, error) {
	record := new(table.TblGuildMemberInfo)
	record.Creator = creator
	record.GuildId = gId
	record.JoinTime = int(base.GLocalizedTime.SecTimeStamp())
	record.MemberUin = uin
	record.Post = post
	record.Vitality = 0

	guildMemberInfoModel := model.GuildMemberInfoModel{Creator: creator, GId: gId}
	_, err := guildMemberInfoModel.AddNewGuildMemberInfo(record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func UpdateGuildMemberInfo(creator, gId int, record *table.TblGuildMemberInfo) (int, error) {
	guildMemberInfoModel := model.GuildMemberInfoModel{Creator: creator, GId: gId}
	return guildMemberInfoModel.UpdateGuildMemberInfo(record)
}

func DeleteGuildMemberInfo(creator, gId int, record *table.TblGuildMemberInfo) (int, error) {
	guildMemberInfoModel := model.GuildMemberInfoModel{Creator: creator, GId: gId}
	return guildMemberInfoModel.DeleteMember(record)
}

func GetGuildMemberByPost(creator, gId int, post string) ([]*table.TblGuildMemberInfo, error) {
	guildMemberInfoModel := model.GuildMemberInfoModel{Creator: creator, GId: gId}
	return guildMemberInfoModel.GetGuildMemberByPost(post)
}

func composeProtoGuildMemberInfo(target *proto.ProtoGuildMemberInfo, member *table.TblGuildMemberInfo, userInfo *table.TblUserInfo, leagueInfo *table.TblLeagueInfo, memberWeeklyVitaliy int, tblUserheadInfo *table.TblUserHeadFrame) (int, error) {
	if target == nil || member == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	target.Post = member.Post
	target.JoinTime = member.JoinTime
	target.WeeklyVitality = memberWeeklyVitaliy
	target.Uin = member.MemberUin

	if userInfo != nil {
		target.Name = userInfo.UserName
		target.Icon = userInfo.Icon
		target.CountryCode = userInfo.ISOCountryCode
		target.Level = userInfo.Level
	}

	if leagueInfo != nil {
		target.LeagueLevel = leagueInfo.CurrentLeagueLevel
	}

	if tblUserheadInfo != nil {
		if userHeadFrameIcon, ok := config.GIconFrameConfig.AttrMap[tblUserheadInfo.CurHeadFrame]; ok {
			target.CurHeadFrame = userHeadFrameIcon.ResourceId
			target.HeadFrameProtypeId = tblUserheadInfo.CurHeadFrame
			target.HeadId = tblUserheadInfo.HeadId
			target.HeadType = tblUserheadInfo.HeadType
		} else {
			target.CurHeadFrame = "iconframe_default"
		}
	}

	return 0, nil
}

func composeProtoGuildMemberList(creator, gId int, memberSlice ...*table.TblGuildMemberInfo) ([]*proto.ProtoGuildMemberInfo, *table.TblUserInfo, error) {
	uinList := make([]int, 0, len(memberSlice))
	for _, member := range memberSlice {
		uinList = append(uinList, member.MemberUin)
	}

	userInfoMap, err := GetUserInfoForClientList(uinList...)
	if err != nil {
		return nil, nil, err
	}

	guildWeeklyVitality, err := GetGuildWeeklyVitalityOfAll(creator, gId)
	if err != nil {
		return nil, nil, err
	}

	var chairman *table.TblUserInfo

	protoMemberList := make([]*proto.ProtoGuildMemberInfo, 0, len(memberSlice))
	for _, member := range memberSlice {
		memWeeklyVitality, ok := guildWeeklyVitality.MemberVitality[member.MemberUin]
		if !ok {
			memWeeklyVitality = 0
		}

		memberUserInfo := userInfoMap[member.MemberUin].UserInfo
		memberLeagueInfo := userInfoMap[member.MemberUin].LeagueInfo

		var userheadInfo table.TblUserHeadFrame
		userheadInfo.CurHeadFrame = userInfoMap[member.MemberUin].HeadFrameProtypeId
		userheadInfo.HeadId = userInfoMap[member.MemberUin].HeadId
		userheadInfo.HeadType = userInfoMap[member.MemberUin].HeadType

		protoMemberInfo := new(proto.ProtoGuildMemberInfo)
		_, err := composeProtoGuildMemberInfo(protoMemberInfo, member, memberUserInfo, memberLeagueInfo, memWeeklyVitality, &userheadInfo)
		if err != nil {
			base.GLog.Debug("ComposeProtoGuildMemberInfo Error: \nUserInfo:%+v\nLeagueInfo:%+v\n", userInfoMap[member.MemberUin].UserInfo, userInfoMap[member.MemberUin].LeagueInfo)
			return nil, nil, err
		}

		protoMemberList = append(protoMemberList, protoMemberInfo)

		if member.Post == GUILD_POST_CHAIRMAN {
			chairman = memberUserInfo
		}
	}

	return protoMemberList, chairman, nil
}
