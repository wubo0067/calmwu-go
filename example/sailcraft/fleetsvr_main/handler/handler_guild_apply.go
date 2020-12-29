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

func (this *GuildHandler) Apply() (int, error) {
	var reqParams proto.ProtoApplyGuildRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	creator, gId, ok := ConvertGuildIdToUinAndId(reqParams.GuildId)
	if !ok {
		return errorcode.ERROR_CODE_GUILD_ID_FORMAT_WRONG, custom_errors.New("guild id format error")
	}

	locker, err := LockGuild(creator, gId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	UnlockGuild(creator, gId, locker)

	guildInfoModel := model.GuildInfoModel{Creator: creator, Id: gId}
	guildInfo, err := guildInfoModel.GetGuildInfo()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	retCode, err := ValidGuildJoinStatus(reqParams.GuildId, guildInfo, false)
	if err != nil {
		return retCode, err
	}

	if guildInfo.JoinType != GUILD_JOIN_TYPE_NEED_AUDIT {
		return errorcode.ERROR_CODE_GUILD_JOIN_TYPE_NOT_NEED_AUDIT, custom_errors.New("guild join type is not NEED_AUDIT(1)")
	}

	var userInfo table.TblUserInfo
	retCode, err = GetUserInfo(this.Request.Uin, &userInfo)
	if err != nil {
		return retCode, err
	}

	// 联赛等级限制
	if guildInfo.CondLeagueLevel > 0 {
		leagueInfo, err := GetLeagueInfo(this.Request.Uin)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		if leagueInfo == nil {
			return errorcode.ERROR_CODE_GUILD_LEAGUE_LEVEL_NOT_REACH, custom_errors.New("league info is nil")
		}

		if leagueInfo.CurrentLeagueLevel < guildInfo.CondLeagueLevel {
			return errorcode.ERROR_CODE_GUILD_LEAGUE_LEVEL_NOT_REACH, custom_errors.New("user league level[%d] not reach %d", leagueInfo.CurrentLeagueLevel, guildInfo.CondLeagueLevel)
		}
	}

	retCode, err = ValidUserJoinGuildPermission(&userInfo, creator, gId)
	if err != nil {
		return retCode, err
	}

	retCode, err = AddGuildApplyInfo(this.Request.Uin, creator, gId)
	if err != nil {
		return retCode, err
	}

	return 0, nil
}

func (this *GuildHandler) ApplyList() (int, error) {
	var reqParams proto.ProtoGetGuildApplyListRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	creator, id, ok := ConvertGuildIdToUinAndId(reqParams.GuildId)
	if !ok {
		return errorcode.ERROR_CODE_GUILD_ID_FORMAT_WRONG, custom_errors.New("format of guild_id is incorrect")
	}

	allApplyInfo, err := GetAllGuildApplyInfo(creator, id)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	uinSlice := make([]int, 0, len(allApplyInfo))
	for _, applyInfo := range allApplyInfo {
		uinSlice = append(uinSlice, applyInfo.ApplyUin)
	}

	allUserInfo, err := GetMultiUserInfo(uinSlice...)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	allLeagueInfo, err := GetMultiLeagueInfo(uinSlice...)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	userInfoMap := make(map[int]*table.TblUserInfo)
	for _, userInfo := range allUserInfo {
		userInfoMap[userInfo.Uin] = userInfo
	}

	leagueInfoMap := make(map[int]*table.TblLeagueInfo)
	for _, leagueInfo := range allLeagueInfo {
		leagueInfoMap[leagueInfo.Uin] = leagueInfo
	}

	var responseData proto.ProtoGetGuildApplyListResponse
	responseData.ApplyInfoList = make([]*proto.ProtoGuildApplyInfo, 0, len(allApplyInfo))

	for _, applyInfo := range allApplyInfo {
		protoApplyInfo := new(proto.ProtoGuildApplyInfo)
		retCode, err := composeProtoGuildApplyInfo(protoApplyInfo, applyInfo, userInfoMap[applyInfo.ApplyUin], leagueInfoMap[applyInfo.ApplyUin])
		if err != nil {
			return retCode, err
		}

		responseData.ApplyInfoList = append(responseData.ApplyInfoList, protoApplyInfo)
	}

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *GuildHandler) HandleApply() (int, error) {
	var reqParams proto.ProtoHandlerGuildApplyRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if reqParams.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("apply uin is invalid")
	}

	if reqParams.Operation != GUILD_APPLY_OPERATION_ALLOW && reqParams.Operation != GUILD_APPLY_OPERATION_DENY {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("opeartion of guild apply info is illeage")
	}

	var userInfo table.TblUserInfo
	_, err = GetUserInfo(this.Request.Uin, &userInfo)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	creator, id, ok := ConvertGuildIdToUinAndId(userInfo.GuildID)
	if !ok {
		return errorcode.ERROR_CODE_GUILD_PERMISSION_DENIED, custom_errors.New("permission denied:user[%d] is not in guild", this.Request.Uin)
	}

	// 1. 判断申请是否存在 -------- 疑问：需不需要判断公会申请是否存在？
	applyInfo, err := GetGuildApplyInfo(creator, id, reqParams.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if applyInfo == nil {
		return errorcode.ERROR_CODE_GUILD_APPLY_NOT_EXIST, custom_errors.New("guild apply info of user[%d] is not exist", reqParams.Uin)
	}

	var responseData proto.ProtoHandlerGuildApplyResponse

	if reqParams.Operation == GUILD_APPLY_OPERATION_DENY {
		retCode, err := DeleteGuildApplyInfo(creator, id, applyInfo.ApplyUin)
		if err != nil {
			return retCode, err
		}

		responseData.Operation = GUILD_APPLY_OPERATION_DENY
		responseData.MessageList = make([]*proto.ProtoMessageInfo, 0)
		this.Response.ResData.Params = responseData

		return 0, nil
	}

	guildLocker, err := LockGuild(creator, id)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	defer UnlockGuild(creator, id, guildLocker)

	// 2. 判断权限
	memberInfo, err := GetGuildMemberInfo(creator, id, this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if GuildPostSort[memberInfo.Post] > GuildPostSort[GUILD_POST_VICE_CHAIRMAN] {
		return errorcode.ERROR_CODE_GUILD_PERMISSION_DENIED, custom_errors.New("permission denied: user[%d] can not handler guild apply", this.Request.Uin)
	}

	// 3. 判断退出公会冷却时间
	leaveInfo, err := GetGuildLeaveInfo(reqParams.Uin)
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

	// 3. 判断人数
	locker, err := LockUserGuild(reqParams.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	defer UnLockerUserGuild(reqParams.Uin, locker)

	guildInfoModel := model.GuildInfoModel{Creator: creator, Id: id}
	guildInfo, err := guildInfoModel.GetGuildInfo()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if guildInfo == nil {
		return errorcode.ERROR_CODE_GUILD_ID_NOT_EXIST, custom_errors.New("guild[%s] not exist", userInfo.GuildID)
	}

	if guildInfo.MemberCount >= config.GGlobalConfig.Guild.MaxMemberCount {
		return errorcode.ERROR_CODE_GUILD_MEMBER_FULL, custom_errors.New("member of guild is full.")
	}

	guildVitality, err := GetGuildVitality(userInfo.GuildID)

	responseData.GuildInfo = new(proto.ProtoGuildInfo)
	retCode, err := composeProtoGuildInfo(responseData.GuildInfo, guildInfo, nil, guildVitality)
	if err != nil {
		return retCode, err
	}

	var applyUserInfo table.TblUserInfo
	retCode, err = GetUserInfo(reqParams.Uin, &applyUserInfo)
	if err != nil {
		return retCode, err
	}

	_, _, ok = ConvertGuildIdToUinAndId(applyUserInfo.GuildID)
	if ok {
		if applyUserInfo.GuildID == userInfo.GuildID {
			return errorcode.ERROR_CODE_GUILD_ALREADY_IN_THIS_GUILD, custom_errors.New("user[%d] is already in this guild[%s]", applyUserInfo.Uin, applyUserInfo.GuildID)
		}

		return errorcode.ERROR_CODE_GUILD_ALREADY_IN_GUILD, custom_errors.New("user[%d] is already in guild[%s]", applyUserInfo.Uin, applyUserInfo.GuildID)
	}

	applyUserMemberInfo, retCode, err := AddNewGuildMember(guildInfo, &applyUserInfo, GUILD_POST_MEMBER)
	if err != nil {
		return retCode, err
	}

	leagueInfo, err := GetLeagueInfo(applyUserInfo.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if leagueInfo == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("league info of user[%d] is not exist", applyUserInfo.Uin)
	}

	tblUserheadInfo, err := GetUserHeadFrame(applyUserInfo.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	var protoGuildmemberInfo proto.ProtoGuildMemberInfo
	composeProtoGuildMemberInfo(&protoGuildmemberInfo, applyUserMemberInfo, &applyUserInfo, leagueInfo, 0, tblUserheadInfo)

	responseData.GuildMembers = append(responseData.GuildMembers, &protoGuildmemberInfo)

	protoMessageInfo, err := AddMemberPostChangedMessage(userInfo.GuildID, this.Request.Uin, &userInfo, &applyUserInfo, GUILD_POST_MEMBER, MEMBER_POST_OPERATION_APPLY_ALLOW)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	retCode, err = DeleteGuildApplyInfo(creator, id, applyInfo.ApplyUin)
	if err != nil {
		return retCode, err
	}

	responseData.MessageList = append(responseData.MessageList, protoMessageInfo)

	this.Response.ResData.Params = responseData

	return 0, nil
}

func AddGuildApplyInfo(applyUin int, creator int, id int) (int, error) {

	applyInfoModel := model.GuildApplyInfoModel{CreatorUin: creator, Id: id}

	// 1. 判断申请信息是否存在
	existsApply, err := applyInfoModel.GetApplyInfo(applyUin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	if existsApply != nil {
		return errorcode.ERROR_CODE_GUILD_APPLY_REPEAT, custom_errors.New("apply join guild repeat")
	}

	// 2. 先添加
	var applyInfo table.TblGuildApplyInfo
	applyInfo.ApplyTime = int(base.GLocalizedTime.SecTimeStamp())
	applyInfo.ApplyUin = applyUin
	applyInfo.GuildId = FormatGuildId(creator, id)

	applyCount, err := applyInfoModel.AddApplyInfo(&applyInfo)
	if err != nil {
		return 0, err
	}

	// 3. 再删除
	delCount := applyCount - config.GGlobalConfig.Guild.LimitGuildApply
	base.GLog.Debug("Apply---Limit: %d Current:%d Delete:%d", config.GGlobalConfig.Guild.LimitGuildApply, applyCount, delCount)
	if delCount > 0 {
		retCode, err := applyInfoModel.DeleteOldestApplyInfo(delCount)
		if err != nil {
			return retCode, err
		}
	}

	return 0, nil
}

func GetGuildApplyInfo(creator int, id int, uin int) (*table.TblGuildApplyInfo, error) {
	applyInfoModel := model.GuildApplyInfoModel{CreatorUin: creator, Id: id}
	return applyInfoModel.GetApplyInfo(uin)
}

func GetAllGuildApplyInfo(creator int, id int) ([]*table.TblGuildApplyInfo, error) {
	applyInfoModel := model.GuildApplyInfoModel{CreatorUin: creator, Id: id}
	applyMap, err := applyInfoModel.GetAllApplyInfo()
	if err != nil {
		return nil, err
	}

	applyInfoArr := make([]*table.TblGuildApplyInfo, 0, len(applyMap))
	for _, v := range applyMap {
		applyInfoArr = append(applyInfoArr, v)
	}

	return applyInfoArr, nil
}

func DeleteAllApplyInfo(creator int, id int) (int, error) {
	applyInfoModel := model.GuildApplyInfoModel{CreatorUin: creator, Id: id}
	return applyInfoModel.DeleteAllApplyInfo()
}

func DeleteGuildApplyInfo(creator int, id int, uin int) (int, error) {
	applyInfoModel := model.GuildApplyInfoModel{CreatorUin: creator, Id: id}
	return applyInfoModel.DeleteApplyInfo(uin)
}

func composeProtoGuildApplyInfo(target *proto.ProtoGuildApplyInfo, applyInfo *table.TblGuildApplyInfo, userInfo *table.TblUserInfo, leagueInfo *table.TblLeagueInfo) (int, error) {
	if target == nil || applyInfo == nil || userInfo == nil || leagueInfo == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	target.Uin = applyInfo.ApplyUin
	target.ApplyTime = applyInfo.ApplyTime
	target.Level = userInfo.Level
	target.LeagueLevel = leagueInfo.CurrentLeagueLevel
	target.Name = userInfo.UserName
	target.Icon = userInfo.Icon
	target.CountryCode = userInfo.ISOCountryCode

	return 0, nil
}
