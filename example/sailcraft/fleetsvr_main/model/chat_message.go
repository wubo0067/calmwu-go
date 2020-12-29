package model

import (
	"encoding/json"
	"fmt"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/table"
)

const (
	CACHED_KEY_CHAT_MESSAGE = "chat_message"
)

type ChatMessageModel struct {
	Channel string
	PoolId  int
}

func (this *ChatMessageModel) CachedKey() string {
	return fmt.Sprintf("%s.%s.%d", CACHED_KEY_CHAT_MESSAGE, this.Channel, this.PoolId)
}

func (this *ChatMessageModel) Validate() (int, error) {
	if this.Channel == "" {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("channel is empty")
	}

	if this.PoolId <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("channel pool id is invalid")
	}

	return 0, nil
}

func (this *ChatMessageModel) AddMessage(record *table.TblChatMessage) (int, error) {
	_, err := this.Validate()
	if err != nil {
		return 0, err
	}

	if record == nil {
		return 0, custom_errors.New("record is empty")
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return 0, custom_errors.New("redisMgr is nil")
	}

	data, err := json.Marshal(record)
	if err != nil {
		return 0, err
	}

	elemCount, err := redisMgr.ListRPushVariable(redisKey, string(data))
	if err != nil {
		return 0, err
	}

	return elemCount, nil
}

func (this *ChatMessageModel) GetAllMessages() ([]*table.TblChatMessage, error) {
	_, err := this.Validate()
	if err != nil {
		return nil, err
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return nil, custom_errors.New("redisMgr is nil")
	}

	dataList, err := redisMgr.ListGet(redisKey)
	if err != nil {
		return nil, err
	}

	records := make([]*table.TblChatMessage, 0, len(dataList))
	for _, data := range dataList {
		if data == "" {
			continue
		}

		chatMessage := new(table.TblChatMessage)
		err = json.Unmarshal([]byte(data), chatMessage)
		if err != nil {
			continue
		}

		records = append(records, chatMessage)
	}

	return records, nil
}

func (this *ChatMessageModel) DeleteOldestMessage(keepCount int) (int, error) {
	retCode, err := this.Validate()
	if err != nil {
		return retCode, err
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisMgr is nil")
	}

	err = redisMgr.ListTrim(redisKey, -keepCount, -1)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *ChatMessageModel) Delete() (int, error) {
	retCode, err := this.Validate()
	if err != nil {
		return retCode, err
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisMgr is nil")
	}

	err = redisMgr.DelKey(redisKey)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *ChatMessageModel) MessageCount() (int, error) {
	retCode, err := this.Validate()
	if err != nil {
		return retCode, err
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisMgr is nil")
	}

	count, err := redisMgr.ListLen(redisKey)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return count, nil
}
