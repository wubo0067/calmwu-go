package handler

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sailcraft/base"
	"sailcraft/dataaccess/redistool"
	"sailcraft/fleetsvr_main/config"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/model"
	"sailcraft/fleetsvr_main/proto"
	"sailcraft/fleetsvr_main/table"
	"sort"
)

const (
	LOCKER_GUILD_FRIGATE_SHIP = "locker.guild_frigate_ship"
)

type GuildFrigateShipFeedMessageContent struct {
	Exp int `json:"Exp"`
}

func (this *GuildHandler) FrigateShipInfoWithoutMessage() (int, error) {
	var reqParams proto.ProtoGetGuildFrigateShipWithoutMessageRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	creator, id, ok := ConvertGuildIdToUinAndId(reqParams.GuildId)
	if !ok {
		return errorcode.ERROR_CODE_GUILD_ID_FORMAT_WRONG, custom_errors.New("guild id[\"%s\"] format error", reqParams.GuildId)
	}

	tblFrigateShip, err := GetGuildFrigateShip(id, creator)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	var responseData proto.ProtoGetGuildFrigateShipWithoutMessageResponse
	retCode, err := composeProtoGuildFrigateShipInfo(&responseData.FrigateShipInfo, tblFrigateShip, -1)
	if err != nil {
		return retCode, err
	}

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *GuildHandler) FrigateShipInfo() (int, error) {
	var reqParams proto.ProtoGetGuildFrigateShipRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	creator, id, ok := ConvertGuildIdToUinAndId(reqParams.GuildId)
	if !ok {
		return errorcode.ERROR_CODE_GUILD_ID_FORMAT_WRONG, custom_errors.New("guild id[\"%s\"] format error", reqParams.GuildId)
	}

	tblFrigateShip, err := GetGuildFrigateShip(id, creator)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	var responseData proto.ProtoGetGuildFrigateShipResponse
	retCode, err := composeProtoGuildFrigateShipInfo(&responseData.FrigateShipInfo, tblFrigateShip, -1)
	if err != nil {
		return retCode, err
	}

	channel := GetGuildFrigateShipChannel(creator, id)
	protoMessageList, err := GetAllProtoMessages(channel)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	sort.Slice(protoMessageList, func(i, j int) bool { return protoMessageList[i].SendTime < protoMessageList[j].SendTime })

	if len(protoMessageList) > 0 {
		responseData.MessageList = append(responseData.MessageList, protoMessageList...)
	} else {
		responseData.MessageList = make([]*proto.ProtoMessageInfo, 0)
	}

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *GuildHandler) FeedFrigateShip() (int, error) {
	var reqParams proto.ProtoFeedGuildFrigateShipRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	reqProps := make([]*proto.ProtoPropUseInfo, 0, len(reqParams.Props))
	for i, _ := range reqParams.Props {
		if reqParams.Props[i].Count > 0 {
			reqProps = append(reqProps, &(reqParams.Props[i]))
		}
	}

	if len(reqProps) <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("props is empty")
	}

	// 1. 判断用户是否在公会中
	var userInfo table.TblUserInfo
	retCode, err := GetUserInfo(this.Request.Uin, &userInfo)
	if err != nil {
		return retCode, err
	}

	creator, gId, ok := ConvertGuildIdToUinAndId(userInfo.GuildID)
	if !ok {
		return errorcode.ERROR_CODE_GUILD_NOT_IN_GUILD, custom_errors.New("user is not in guild")
	}

	// 2. 判断道具是否足够
	protypeMap := make(map[int]*config.PropProtype)
	protypeIds := make([]int, 0, len(reqProps))
	for _, protoProp := range reqProps {
		protype, ok := config.GPropConfig.AttrMap[protoProp.ProtypeId]
		if !ok {
			return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("protype id not exists")
		}

		if protype.PropType != config.PROP_TYPE_GUILD_FRIGATE_EXP {
			return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("prop[%d] is can not feed guild frigate ship", protype.Id)
		}

		protypeMap[protype.Id] = protype
		protypeIds = append(protypeIds, protype.Id)
	}

	props, err := GetMultiPropsByProtypeId(this.Request.Uin, protypeIds...)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if len(props) < len(protypeIds) {
		return errorcode.ERROR_CODE_GUILD_PROP_NOT_ENOUGH, custom_errors.New("prop num is not enough")
	}

	propMap := make(map[int]*table.TblPropInfo)
	for _, prop := range props {
		propMap[prop.ProtypeId] = prop
	}

	for _, protoProp := range reqProps {
		prop, ok := propMap[protoProp.ProtypeId]
		if !ok {
			return errorcode.ERROR_CODE_GUILD_PROP_NOT_ENOUGH, custom_errors.New("prop[%d] not exists", protoProp.ProtypeId)
		}

		if prop.PropNum < protoProp.Count {
			return errorcode.ERROR_CODE_GUILD_PROP_NOT_ENOUGH, custom_errors.New("prop[%d] num is not enough", protoProp.ProtypeId)
		}
	}

	locker, err := LockGuildFrigateShip(creator, gId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	defer UnlockGuildFrigateShip(creator, gId, locker)

	// 3. 更新公会护卫舰
	tblFrigateShip, err := GetGuildFrigateShip(gId, creator)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	incrExp := 0
	var effect config.PropGuildFriExpEffect
	for _, protoProp := range reqProps {
		protype := protypeMap[protoProp.ProtypeId]
		err = config.GPropConfig.DecodeEffect(protype, &effect)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
		incrExp += protoProp.Count * effect.Exp
	}

	resValue := config.GGuildFrigateConfig.AttrArr[tblFrigateShip.Level-1].FeedbackRatio * float32(incrExp)

	tblFrigateShip.Exp += incrExp
	newLevel, restExp := config.GGuildFrigateConfig.LevelExp(tblFrigateShip.Level, tblFrigateShip.Exp)
	oldLevel := tblFrigateShip.Level
	tblFrigateShip.Level = newLevel
	tblFrigateShip.Exp = restExp

	retCode, err = UpdateGuildFrigateShip(gId, creator, tblFrigateShip)
	if err != nil {
		return retCode, err
	}

	// 4. 计算物资回馈（奖励）
	r := rand.New(rand.NewSource(base.GLocalizedTime.NSecTimeStamp()))
	resCount := r.Intn(4) + 1
	resRatio := make([]float32, 0, resCount)
	totalRatio := float32(0.0)

	for i := 0; i < resCount; i++ {
		ratio := float32(r.Intn(100) + 1)
		totalRatio += ratio
		resRatio = append(resRatio, ratio)
	}

	for i := 0; i < resCount; i++ {
		resRatio[i] = resValue * float32(resRatio[i]) / float32(totalRatio)
	}

	// 5. 返回公会护卫舰信息以及道具消耗信息
	var responseData proto.ProtoFeedGuildFrigateShipResponse
	retCode, err = composeProtoGuildFrigateShipInfo(&responseData.FrigateShipInfo, tblFrigateShip, oldLevel)
	if err != nil {
		return retCode, err
	}

	responseData.Cost.BattleShipCards = make([]*proto.ProtoBattleShipCardItem, 0)
	responseData.Cost.ResourceItems = make([]*proto.ProtoResourceItem, 0)
	responseData.Cost.PropItems = make([]*proto.ProtoPropItem, 0, len(protypeIds))
	for _, protoProp := range reqProps {
		item := new(proto.ProtoPropItem)
		item.Count = protoProp.Count
		item.ProtypeId = protoProp.ProtypeId
		item.CountType = config.COUNT_TYPE_CONST
		responseData.Cost.PropItems = append(responseData.Cost.PropItems, item)
	}

	responseData.Rewards.BattleShipCards = make([]*proto.ProtoBattleShipCardItem, 0)
	responseData.Rewards.PropItems = make([]*proto.ProtoPropItem, 0)
	responseData.Rewards.ResourceItems = make([]*proto.ProtoResourceItem, 0, resCount)

	// 金币
	n := r.Intn(24)
	index := 0
	if resCount > 0 && n < (resCount*6) {
		count := int(math.Max(1.0, float64(resRatio[index]*config.GGlobalConfig.Guild.ValueRatio.Gold)))

		protoResItem := new(proto.ProtoResourceItem)
		protoResItem.Type = config.RESOURCE_ITEM_TYPE_GOLD
		protoResItem.CountType = config.COUNT_TYPE_CONST
		protoResItem.Count = count
		responseData.Rewards.ResourceItems = append(responseData.Rewards.ResourceItems, protoResItem)

		resCount--
		index++
	}

	// 木材
	n = r.Intn(24)
	if resCount > 0 && n < (resCount*8) {
		count := int(math.Max(1.0, float64(resRatio[index]*config.GGlobalConfig.Guild.ValueRatio.Wood)))

		protoResItem := new(proto.ProtoResourceItem)
		protoResItem.Type = config.RESOURCE_ITEM_TYPE_WOOD
		protoResItem.CountType = config.COUNT_TYPE_CONST
		protoResItem.Count = count
		responseData.Rewards.ResourceItems = append(responseData.Rewards.ResourceItems, protoResItem)

		resCount--
		index++
	}

	// 铁矿
	n = r.Intn(24)
	if resCount > 0 && n < (resCount*16) {
		count := int(math.Max(1.0, float64(resRatio[index]*config.GGlobalConfig.Guild.ValueRatio.Iron)))

		protoResItem := new(proto.ProtoResourceItem)
		protoResItem.Type = config.RESOURCE_ITEM_TYPE_IRON
		protoResItem.CountType = config.COUNT_TYPE_CONST
		protoResItem.Count = count
		responseData.Rewards.ResourceItems = append(responseData.Rewards.ResourceItems, protoResItem)

		resCount--
		index++
	}

	// 石材
	if resCount > 0 {
		count := int(math.Max(1.0, float64(resRatio[index]*config.GGlobalConfig.Guild.ValueRatio.Stone)))

		protoResItem := new(proto.ProtoResourceItem)
		protoResItem.Type = config.RESOURCE_ITEM_TYPE_STONE
		protoResItem.CountType = config.COUNT_TYPE_CONST
		protoResItem.Count = count
		responseData.Rewards.ResourceItems = append(responseData.Rewards.ResourceItems, protoResItem)

		resCount--
		index++
	}

	channel := GetGuildFrigateShipChannel(creator, gId)
	protoMessage, err := AddGuildFrigateShipFeedMessage(channel, this.Request.Uin, incrExp)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	responseData.MessageList = append(responseData.MessageList, protoMessage)

	this.Response.ResData.Params = responseData

	return 0, nil
}

func CreateFrigateShipInfo(gId, creator int) (*table.TblGuildFrigateShip, error) {
	firgateShip := newDefaultGuildFrigateShip(gId, creator)

	guildFrigateShipModel := model.GuildFrigateShipModel{Id: gId, Creator: creator}

	_, err := guildFrigateShipModel.Insert(firgateShip)
	if err != nil {
		return nil, err
	}

	return firgateShip, nil
}

func GetGuildFrigateShip(gId, creator int) (*table.TblGuildFrigateShip, error) {
	guildFrigateShipModel := model.GuildFrigateShipModel{Id: gId, Creator: creator}
	frigateShip, err := guildFrigateShipModel.Query()
	if err != nil {
		return nil, err
	}

	if frigateShip == nil {
		frigateShip = newDefaultGuildFrigateShip(gId, creator)
	}

	return frigateShip, nil
}

func UpdateGuildFrigateShip(gId, creator int, record *table.TblGuildFrigateShip) (int, error) {
	guildFrigateShipModel := model.GuildFrigateShipModel{Id: gId, Creator: creator}
	if record.Id > 0 {
		retCode, err := guildFrigateShipModel.Update(record)
		if err != nil {
			return retCode, err
		}
	} else {
		retCode, err := guildFrigateShipModel.Insert(record)
		if err != nil {
			return retCode, err
		}
	}

	return 0, nil
}

func DeleteGuildFrigateShip(gId, creator int) (int, error) {
	guildFrigateShipModel := model.GuildFrigateShipModel{Id: gId, Creator: creator}
	retCode, err := guildFrigateShipModel.Delete()
	if err != nil {
		return retCode, err
	}

	channel := GetGuildFrigateShipChannel(creator, gId)
	retCode, err = DeleteChannel(channel)
	if err != nil {
		return retCode, err
	}

	return 0, nil
}

func newDefaultGuildFrigateShip(gId, creator int) *table.TblGuildFrigateShip {
	frigateShip := new(table.TblGuildFrigateShip)
	frigateShip.Exp = 0
	frigateShip.Level = 1
	frigateShip.GuildId = gId
	frigateShip.GuildCreator = creator
	frigateShip.ProtypeId = config.GGuildFrigateConfig.AttrArr[0].Id

	return frigateShip
}

func composeProtoGuildFrigateShipInfo(target *proto.ProtoGuildFrigateShipInfo, data *table.TblGuildFrigateShip, oldLevel int) (int, error) {
	if target == nil || data == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	target.GuildId = FormatGuildId(data.GuildCreator, data.GuildId)
	target.Exp = data.Exp
	target.Level = data.Level
	target.ProtypeId = data.ProtypeId
	if oldLevel <= 0 {
		target.OldLevel = target.Level
	} else {
		target.OldLevel = oldLevel
	}

	return 0, nil
}

func LockGuildFrigateShip(creator, gId int) (string, error) {
	key := fmt.Sprintf("%s.%s", LOCKER_GUILD_FRIGATE_SHIP, FormatGuildId(creator, gId))
	return redistool.SpinLockWithFingerPoint(key, 0)
}

func UnlockGuildFrigateShip(creator, gId int, value string) {
	key := fmt.Sprintf("%s.%s", LOCKER_GUILD_FRIGATE_SHIP, FormatGuildId(creator, gId))
	err := redistool.UnLock(key, value)
	if err != nil {
		base.GLog.Error("unlock %s failed! reason[%s]", err)
	}
}

func GetGuildFrigateShipChannel(creator, gId int) string {
	return fmt.Sprintf("%s.frigate_ship", FormatGuildId(creator, gId))
}

func AddGuildFrigateShipFeedMessage(channel string, uin, exp int) (*proto.ProtoMessageInfo, error) {
	protoMessageContent := new(proto.ProtoFeedGuildFrigateShipMessage)
	protoMessageContent.Exp = exp

	data, err := json.Marshal(protoMessageContent)
	if err != nil {
		return nil, err
	}

	return AddMessage(channel, uin, string(data), CHAT_MESSAGE_TYPE_GUILD_FRIGATE_SHIP_FEED)
}
