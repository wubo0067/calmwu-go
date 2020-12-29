package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sailcraft/base"
)

type ChatMessagePool struct {
	Id    int `json:"id"`
	Limit int `json:"limit"`
}

type ChatMessageType struct {
	Type   string `json:"type"`
	PoolId int    `json:"pool_id"`
}

type ChatMessageConfigBase struct {
	Pools     []*ChatMessagePool `json:"pool"`
	ChatTypes []*ChatMessageType `json:"chat_type"`
}

type ChatMessageConfig struct {
	ChatMessageConfigBase
	PoolMap     map[int]*ChatMessagePool
	ChatTypeMap map[string]*ChatMessageType
}

var (
	GChatMessageConfig = new(ChatMessageConfig)
)

func (this *ChatMessageConfig) Init(configFile string) error {
	base.GLog.Debug("LoadConfig [%s]", configFile)

	hFile, err := os.Open(configFile)
	if err != nil {
		base.GLog.Error("open file %s failed err %s \n", configFile, err.Error())
		return err
	}
	defer hFile.Close()

	data, err := ioutil.ReadAll(hFile)
	if err != nil {
		base.GLog.Error("read file %s failed err %s \n", configFile, err.Error())
		return err
	}

	err = json.Unmarshal(data, &this.ChatMessageConfigBase)
	if err != nil {
		return err
	}

	this.PoolMap = make(map[int]*ChatMessagePool)
	this.ChatTypeMap = make(map[string]*ChatMessageType)

	for _, pool := range this.Pools {
		this.PoolMap[pool.Id] = pool
	}

	for _, chattype := range this.ChatTypes {
		this.ChatTypeMap[chattype.Type] = chattype
	}

	base.GLog.Debug("chat message config data is [%+v]", *this)

	return nil
}
