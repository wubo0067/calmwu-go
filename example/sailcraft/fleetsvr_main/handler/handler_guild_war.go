package handler

import (
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
)

const (
	GUILD_WAR_STATUS_FINISHED = 0 // 已结束
	GUILD_WAR_STATUS_ONGOING  = 1 // 正在进行中

	GUILD_WAR_TYPE_FIRST  = 1 // 上半场
	GUILD_WAR_TYPE_SECOND = 2 // 下半场

	GUILD_WAR_PHASE_STOP           = 1
	GUILD_WAR_PHASE_FIRST_ONGOING  = 2
	GUILD_WAR_PHASE_FIRST_END      = 3
	GUILD_WAR_PHASE_SECOND_ONGOING = 4
	GUILD_WAR_PHASE_SECOND_END     = 5
	GUILD_WAR_PHASE_SETTLING       = 6

	GUILD_WAR_WRITER_LOCKER = "guild.war.writer.locker"
)

type GuildWarHandler struct {
	handlerbase.WebHandler
}

func (this *GuildWarHandler) Info() (int, error) {
	tblWarInfo, err := GetGuildWar()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	var responseData proto.ProtoGetGuildWarInfoResponse
	composeProtoGuildWarInfo(&responseData.GuildWarInfo, tblWarInfo)

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *GuildWarHandler) MemberRank() (int, error) {
	var reqParams proto.ProtoGuildWarMemberRankRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	creator, gId, ok := ConvertGuildIdToUinAndId(reqParams.GuildId)
	if !ok {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("guild id format is wrong")
	}

	guildWarMembers, err := GetAllWarMemberOfGuild(creator, gId, reqParams.WarId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	var responseData proto.ProtoGuildWarMemberRankResponse
	responseData.MemberBattleInfoList, err = composeProtoGuildMemberBattleInfoList(guildWarMembers...)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *GuildWarHandler) GuildWarBattleSettle() (int, error) {
	var reqParams proto.ProtoGuildWarBattleSettleRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	creator, gid, ok := ConvertGuildIdToUinAndId(reqParams.GuildId)
	if !ok {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("guild id format error")
	}

	warMemberInfo, err := GetGuildWarMember(creator, gid, reqParams.WarId, this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	// 判断玩家剩余战斗次数
	if warMemberInfo.RestBattleTimes <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("guild battle times is not enough")
	}

	pointsIncr := 0
	winStreak := 0
	switch reqParams.BattleResult {
	case BATTLE_RESULT_DRAW:
		pointsIncr = config.GGlobalConfig.Guild.BattleConfig.DrawScore
	case BATTLE_RESULT_FAILED:
		pointsIncr = config.GGlobalConfig.Guild.BattleConfig.LoseScore
	case BATTLE_RESULT_SUCCESS:
		battledTimes := len(warMemberInfo.BattleRecordList)
		if battledTimes > 0 && warMemberInfo.BattleRecordList[battledTimes-1].BattleResult == BATTLE_RESULT_SUCCESS {
			winStreak = 1
			pointsIncr = config.GGlobalConfig.Guild.BattleConfig.WinStreakScore
		} else {
			pointsIncr = config.GGlobalConfig.Guild.BattleConfig.WinScore
		}
	default:
		return 0, nil
	}

	// 计算积分
	battleRecord := new(table.TblGuildMemberBattleRecord)
	battleRecord.ScoreIncr = pointsIncr
	battleRecord.BattleResult = reqParams.BattleResult

	// 1. 修改成员积分
	warMemberInfo.RestBattleTimes--
	warMemberInfo.BattleRecordList = append(warMemberInfo.BattleRecordList, battleRecord)
	warMemberInfo.Score += pointsIncr
	retCode, err := UpdateGuildWarMember(creator, gid, reqParams.WarId, warMemberInfo)
	if err != nil {
		return retCode, err
	}

	// 2. 增加公会积分
	newPoints := 0
	if pointsIncr > 0 {
		newPoints, err = IncreaseGuildWarPoints(reqParams.WarId, creator, gid, pointsIncr)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
	}

	// 3. 修改公会战斗信息
	locker, err := LockGuildWarForSingleGuild(creator, gid)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	defer UnlockGuildWarForSingleGuild(creator, gid, locker)

	guildWarInfo, err := GetGuildWarInfo(creator, gid, reqParams.WarId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	switch reqParams.BattleResult {
	case BATTLE_RESULT_FAILED:
		guildWarInfo.LoseTimes++
	case BATTLE_RESULT_SUCCESS:
		guildWarInfo.WinTimes++
	}

	guildWarInfo.TotalTimes++
	if newPoints != 0 {
		guildWarInfo.Score = newPoints
	}
	if warMemberInfo.RestBattleTimes == (config.GGlobalConfig.Guild.BattleConfig.PlayerBattleTimes - 1) {
		guildWarInfo.UserCount++
	}

	retCode, err = UpdateGuildWarInfo(creator, gid, reqParams.WarId, guildWarInfo)
	if err != nil {
		return retCode, err
	}

	var responseData proto.ProtoGuildWarBattleSettleResponse
	responseData.ScoreDelta = pointsIncr
	responseData.Score = warMemberInfo.Score
	responseData.GuildScore = guildWarInfo.Score
	responseData.WinStreak = winStreak
	this.Response.ResData.Params = responseData

	return 0, nil
}

func GetGuildWar() (*table.TblGuildWar, error) {
	guildWarModel := model.GuildWarModel{}
	guildWar, err := guildWarModel.Query()
	if err != nil {
		return nil, err
	}

	if guildWar == nil {
		guildWar.GuildWardId = 1
		guildWar.Phase = GUILD_WAR_PHASE_STOP
		guildWar.RewardGroupId = 1
		guildWar.StartTime = int(base.GLocalizedTime.SecTimeStamp() + base.SecondsPerDay*3)
	}

	//TODO 公会战暂未开放
	guildWar.Phase = GUILD_WAR_PHASE_STOP

	return guildWar, nil
}

func composeProtoGuildWarInfo(target *proto.ProtoGuildWarInfo, data *table.TblGuildWar) (int, error) {
	if target == nil || data == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	target.GuildWardId = data.GuildWardId
	target.RewardGroupId = data.RewardGroupId
	target.GuildWarPhase = data.Phase
	switch data.Phase {
	case GUILD_WAR_PHASE_STOP:
		y, m, d := base.GLocalizedTime.UnixDate(int64(data.StartTime)+base.SecondsPerDay*3, 0)
		dayBegin, _ := base.GLocalizedTime.Clock(y, int(m), d, 0, 0, 1)
		ts := int(dayBegin.Unix()) + config.GGlobalConfig.Guild.BattleConfig.FirstHalfInDay
		target.RestTime = ts - int(base.GLocalizedTime.SecTimeStamp())
	case GUILD_WAR_PHASE_FIRST_ONGOING:
		target.RestTime = (data.StartTime + config.GGlobalConfig.Guild.BattleConfig.Duration) - int(base.GLocalizedTime.SecTimeStamp())
	case GUILD_WAR_PHASE_FIRST_END:
		y, m, d := base.GLocalizedTime.UnixDate(int64(data.StartTime)+base.SecondsPerDay*3, 0)
		dayBegin, _ := base.GLocalizedTime.Clock(y, int(m), d, 0, 0, 1)
		ts := int(dayBegin.Unix()) + config.GGlobalConfig.Guild.BattleConfig.SecondHalfInDay
		target.RestTime = ts - int(base.GLocalizedTime.SecTimeStamp())
	case GUILD_WAR_PHASE_SECOND_ONGOING:
		target.RestTime = (data.StartTime + config.GGlobalConfig.Guild.BattleConfig.Duration) - int(base.GLocalizedTime.SecTimeStamp())
	case GUILD_WAR_PHASE_SECOND_END:
		target.RestTime = 0
	case GUILD_WAR_PHASE_SETTLING:
		target.RestTime = 0
	}

	return 0, nil
}

func LockGuildWarForSingleGuild(creator, gId int) (string, error) {
	k := fmt.Sprintf("%s.%d.%d", GUILD_WAR_WRITER_LOCKER, creator, gId)
	return redistool.SpinLockWithFingerPoint(k, 0)
}

func UnlockGuildWarForSingleGuild(creator int, gId int, value string) {
	k := fmt.Sprintf("%s.%d.%d", GUILD_WAR_WRITER_LOCKER, creator, gId)
	err := redistool.UnLock(k, value)
	if err != nil {
		base.GLog.Error("unlock guild war for single guild (creatorUin:%d, gId:%d, value:%s) error[%v]", creator, gId, value, err)
	}
}
