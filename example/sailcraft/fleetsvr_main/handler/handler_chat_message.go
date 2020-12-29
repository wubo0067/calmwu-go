package handler

import (
	"sailcraft/base"
	"sailcraft/fleetsvr_main/config"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/handlerbase"
	"sailcraft/fleetsvr_main/model"
	"sailcraft/fleetsvr_main/proto"
	"sailcraft/fleetsvr_main/table"
	"sort"
)

const (
	CHAT_MESSAGE_TYPE_NORMAL                  = "chat"                     // 聊天
	CHAT_MESSAGE_TYPE_MEMBER_COUNT_CHANGED    = "member_count_changed"     // 成员增减
	CHAT_MESSAGE_TYPE_MEMBER_POST_CHANGED     = "member_post_changed"      // 职级变动
	CHAT_MESSAGE_TYPE_CHAIRMAN_TRANSFER       = "member_chairman_transfer" // 会长转让
	CHAT_MESSAGE_TYPE_GUILD_FRIGATE_SHIP_FEED = "guild_frigate_ship_feed"  // 公会护卫舰培养

	MEMBER_COUNT_OPERATION_JOIN  = 0
	MEMBER_COUNT_OPERATION_LEAVE = 1

	MEMBER_POST_OPERATION_APPLY_ALLOW = 0
	MEMBER_POST_OPERATION_KICK        = 1
	MEMBER_POST_OPERATION_PROMOTE     = 2
	MEMBER_POST_OPERATION_DEMOTE      = 3
)

type MessageHandler struct {
	handlerbase.WebHandler
}

func (this *MessageHandler) List() (int, error) {
	var reqParams proto.ProtoMessageListRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	protoMessageList, err := GetAllProtoMessages(reqParams.Channel)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	sort.Slice(protoMessageList, func(i, j int) bool { return protoMessageList[i].SendTime < protoMessageList[j].SendTime })

	var responseData proto.ProtoMessageListResponse
	responseData.MessageList = protoMessageList[:]

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *MessageHandler) Send() (int, error) {
	var reqParams proto.ProtoSendMessageRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	protoMessage, err := AddMessage(reqParams.Channel, this.Request.Uin, reqParams.Content, CHAT_MESSAGE_TYPE_NORMAL)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	var responseData proto.ProtoSendMessageResponse
	responseData.MessageList = append(responseData.MessageList, protoMessage)

	this.Response.ResData.Params = responseData

	return 0, nil
}

func GetAllProtoMessages(channel string) ([]*proto.ProtoMessageInfo, error) {
	tblMessageList, err := GetAllMessages(channel)
	if err != nil {
		return nil, err
	}

	protoMessageList := make([]*proto.ProtoMessageInfo, 0, len(tblMessageList))
	for _, tblMessage := range tblMessageList {
		protoMessage := new(proto.ProtoMessageInfo)
		composeProtoChatMessage(protoMessage, tblMessage)
		protoMessageList = append(protoMessageList, protoMessage)
	}

	return protoMessageList, nil
}

func GetAllMessages(channel string) ([]*table.TblChatMessage, error) {
	messageList := make([]*table.TblChatMessage, 0)

	chatMessageModel := model.ChatMessageModel{Channel: channel}
	for _, pool := range config.GChatMessageConfig.Pools {
		chatMessageModel.PoolId = pool.Id
		poolMessages, err := chatMessageModel.GetAllMessages()
		if err != nil {
			return nil, err
		}
		messageList = append(messageList, poolMessages...)
	}

	return messageList, nil
}

func AddMessageRecord(channel string, record *table.TblChatMessage) (int, error) {
	if record == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	chatType, ok := config.GChatMessageConfig.ChatTypeMap[record.Type]
	if !ok {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("chat type[%s] is not exist", record.Type)
	}

	pool, ok := config.GChatMessageConfig.PoolMap[chatType.PoolId]
	if !ok {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("pool[%d] is not exist", chatType.PoolId)
	}

	chatMessageModel := model.ChatMessageModel{Channel: channel, PoolId: pool.Id}
	msgCount, err := chatMessageModel.AddMessage(record)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if msgCount > pool.Limit {
		retCode, err := chatMessageModel.DeleteOldestMessage(pool.Limit)
		if err != nil {
			return retCode, err
		}
	}

	return 0, nil
}

func AddMessage(channel string, uin int, content, messageType string) (*proto.ProtoMessageInfo, error) {
	record := new(table.TblChatMessage)
	record.Message = content
	record.SendTime = int(base.GLocalizedTime.SecTimeStamp())
	record.Type = messageType
	record.Uin = uin

	_, err := AddMessageRecord(channel, record)
	if err != nil {
		return nil, err
	}

	protoMessage := new(proto.ProtoMessageInfo)
	err = composeProtoChatMessage(protoMessage, record)
	if err != nil {
		return nil, err
	}

	return protoMessage, err
}

func DeleteChannel(channel string) (int, error) {
	chatMessageModel := model.ChatMessageModel{Channel: channel}
	for _, pool := range config.GChatMessageConfig.Pools {
		chatMessageModel.PoolId = pool.Id
		retCode, err := chatMessageModel.Delete()
		if err != nil {
			return retCode, err
		}
	}

	return 0, nil
}

func composeProtoChatMessage(target *proto.ProtoMessageInfo, data *table.TblChatMessage) error {
	if target == nil || data == nil {
		return custom_errors.NullPoint()
	}

	target.Content = data.Message
	target.MessageType = data.Type
	target.SendTime = data.SendTime
	target.Uin = data.Uin

	return nil
}
