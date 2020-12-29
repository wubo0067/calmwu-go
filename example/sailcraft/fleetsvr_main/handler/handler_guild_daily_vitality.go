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

func (this *GuildHandler) GuildTaskInfo() (int, error) {
	dailyVitaliy, dailyVitalityRewards, err := GetGuildDailyVitaliyAndRewardList(this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	mapP2D := make(map[int]*table.TblGuildDailyVitalityReward)
	for _, record := range dailyVitalityRewards {
		mapP2D[record.ProtypeId] = record
	}

	protoRewardList := make([]*proto.ProtoGuildTaskRewardInfo, 0, len(config.GGuildTaskRewardConfig.AttrArr))
	for _, protype := range config.GGuildTaskRewardConfig.AttrArr {
		protoReward := new(proto.ProtoGuildTaskRewardInfo)
		_, err := composeProtoGuildTaskReward(protoReward, protype, mapP2D[protype.Id])
		if err != nil {
			base.GLog.Error("compose ProtoGuildTaskRewardInfo failed[%s]", err)
			continue
		}

		protoRewardList = append(protoRewardList, protoReward)
	}

	t, err := base.GLocalizedTime.TodayClock(23, 59, 59)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	var responseData proto.ProtoGuildTaskInfoResponse
	responseData.Uin = this.Request.Uin
	responseData.Vitality = dailyVitaliy.Vitality
	responseData.RestTime = int(t.Unix() - base.GLocalizedTime.SecTimeStamp())
	responseData.RewardList = protoRewardList

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *GuildHandler) ReceiveVitalityReward() (int, error) {
	var reqParams proto.ProtoReceiveVitalityRewardRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	_, dailyVitalityRewards, err := GetGuildDailyVitaliyAndRewardList(this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	mapP2D := make(map[int]*table.TblGuildDailyVitalityReward)
	for _, record := range dailyVitalityRewards {
		mapP2D[record.ProtypeId] = record
	}

	if rec, ok := mapP2D[reqParams.ProtypeId]; ok {
		switch rec.Status {
		case GUILD_DAILY_VITALITY_REWARD_STATUS_UNCOMPLETED:
			return errorcode.ERROR_CODE_GUILD_TASK_REWARD_VITALITY_NOT_ENOUGH, custom_errors.New("vitality is not enough")
		case GUILD_DAILY_VITALITY_REWARD_STATUS_RECEIVED:
			return errorcode.ERROR_CODE_GUILD_TASK_REWARD_ALREADY_RECEIVED, custom_errors.New("reward is already received")
		}

		rec.Status = GUILD_DAILY_VITALITY_REWARD_STATUS_RECEIVED
		retCode, err := UpdateGuildDailyVitalityReward(this.Request.Uin, rec)
		if err != nil {
			return retCode, err
		}

		protype, ok := config.GGuildTaskRewardConfig.AttrMap[reqParams.ProtypeId]
		if !ok {
			return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("protype not found")
		}

		protoReward := new(proto.ProtoGuildTaskRewardInfo)
		composeProtoGuildTaskReward(protoReward, protype, rec)

		var responseData proto.ProtoReceiveVitalityRewardResponse
		responseData.RewardList = append(responseData.RewardList, protoReward)
		ResourcesConfigToProto(&protype.Reward, &responseData.Rewards)

		this.Response.ResData.Params = responseData

		return 0, nil
	} else {
		return errorcode.ERROR_CODE_GUILD_TASK_REWARD_VITALITY_NOT_ENOUGH, custom_errors.New("not found reward data")
	}
}

func GetGuildDailyVitality(uin int) (*table.TblGuildDailyVitality, error) {
	guildDailyVitalityModel := model.GuildDailyVitalityModel{Uin: uin}
	todayVitality, err := guildDailyVitalityModel.GetVitality()
	if err != nil {
		return nil, err
	}

	if todayVitality == nil {
		todayVitality = new(table.TblGuildDailyVitality)
		todayVitality.Uin = uin
		todayVitality.FreshTime = int(base.GLocalizedTime.SecTimeStamp())
		todayVitality.Vitality = 0
		_, err := guildDailyVitalityModel.AddGuildDialyVitality(todayVitality)
		if err != nil {
			return nil, err
		}

		return todayVitality, err
	} else {
		if !base.GLocalizedTime.IsToday(int64(todayVitality.FreshTime)) {
			guildDailyVitalityRewardModel := model.GuildDailyVitalityRewardModel{Uin: uin}
			_, err = guildDailyVitalityRewardModel.ResetStatus()
			if err != nil {
				return nil, err
			}

			todayVitality.FreshTime = int(base.GLocalizedTime.SecTimeStamp())
			todayVitality.Vitality = 0
			_, err := guildDailyVitalityModel.UpdateGuildDailyVitality(todayVitality)
			if err != nil {
				return nil, err
			}
		}

		return todayVitality, nil
	}
}

func UdpateGuildDailyVitality(uin int, record *table.TblGuildDailyVitality, updateCols ...string) (int, error) {
	guildDailyVitalityModel := model.GuildDailyVitalityModel{Uin: uin}
	return guildDailyVitalityModel.UpdateGuildDailyVitality(record, updateCols...)
}

func GetGuildDailyVitaliyAndRewardList(uin int) (*table.TblGuildDailyVitality, []*table.TblGuildDailyVitalityReward, error) {
	dailyVitality, err := GetGuildDailyVitality(uin)
	if err != nil {
		return nil, nil, err
	}

	guildDailyVitalityRewardModel := model.GuildDailyVitalityRewardModel{Uin: uin}
	rewardSlice, err := guildDailyVitalityRewardModel.GetRewardList()
	if err != nil {
		return nil, nil, err
	}

	return dailyVitality, rewardSlice, nil
}

func UpdateGuildDailyVitalityReward(uin int, record *table.TblGuildDailyVitalityReward) (int, error) {
	guildDailyVitalityRewardModel := model.GuildDailyVitalityRewardModel{Uin: uin}
	return guildDailyVitalityRewardModel.UpdateReward(record)
}

func AddGuildDailyVitalityReward(uin int, record *table.TblGuildDailyVitalityReward) (int, error) {
	guildDailyVitalityRewardModel := model.GuildDailyVitalityRewardModel{Uin: uin}
	return guildDailyVitalityRewardModel.InsertReward(record)
}

func AddMultiGuildDailyVitalityReward(uin int, recordSlice []*table.TblGuildDailyVitalityReward) (int, error) {
	guildDailyVitalityRewardModel := model.GuildDailyVitalityRewardModel{Uin: uin}
	return guildDailyVitalityRewardModel.InsertMultiReward(recordSlice)
}

func UpdateMultiGuildDailyVitalityReward(uin int, recordSlice []*table.TblGuildDailyVitalityReward) (int, error) {
	guildDailyVitalityRewardModel := model.GuildDailyVitalityRewardModel{Uin: uin}
	return guildDailyVitalityRewardModel.UpdateMultiRewards(recordSlice)
}

func UpdateDailyVitalityAndReward(uin int, vitalityRec *table.TblGuildDailyVitality, rewardRecSlice []*table.TblGuildDailyVitalityReward) (int, error) {
	if vitalityRec != nil {
		guildDailyVitalityModel := model.GuildDailyVitalityModel{Uin: uin}
		retCode, err := guildDailyVitalityModel.UpdateGuildDailyVitality(vitalityRec)
		if err != nil {
			return retCode, err
		}
	}

	guildDailyVitalityRewardModel := model.GuildDailyVitalityRewardModel{Uin: uin}
	retCode, err := guildDailyVitalityRewardModel.UpdateMultiRewards(rewardRecSlice)
	if err != nil {
		return retCode, err
	}

	return 0, nil
}
