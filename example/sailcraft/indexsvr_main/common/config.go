/*
 * @Author: calmwu
 * @Date: 2017-09-18 10:40:09
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-09-19 15:36:44
 * @Comment:
 */

// 读取配置文件

package common

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sailcraft/base"
)

type RedisKeyS struct {
	Keys     []string `json:"Keys"`
	KeyMatch string   `json:"KeyMatch"`
}

type IndexSvrConfigS struct {
	RedisAddressLst []string  `json:"RedisAddressLst"`
	BucketCount     int       `json"BucketCount"`
	Auth            string    `json:"Auth"`
	GuildInfo       RedisKeyS `json:"GuildInfo"`
	UserInfo        RedisKeyS `json:"UserInfo"`
	IndexSvrCluster []string  `json:"IndexSvrCluster"`
	DirtyWordFile   string    `json:"DirtyWordFile"`
}

var (
	GConfig *IndexSvrConfigS = nil
)

func init() {
	if GConfig == nil {
		GConfig = new(IndexSvrConfigS)
	}
}

func LoadConfig(configFile, localAddr string) error {
	base.GLog.Debug("LoadConfig [%s]", configFile)

	hFile, err := os.Open(configFile)
	if err != nil {
		base.GLog.Error("open [%s] failed! err[%s]\n", configFile, err.Error())
		return err
	}
	defer hFile.Close()

	data, err := ioutil.ReadAll(hFile)
	if err != nil {
		base.GLog.Error("read [%s] failed! err[%s]\n", configFile, err.Error())
		return err
	}

	err = json.Unmarshal(data, GConfig)
	if err != nil {
		base.GLog.Error("unmarshal [%s] file failed! err[%s]\n", configFile, err.Error())
		return err
	}

	// 从IndexSvrCluster删除自己
	delPos := 0
	for index, svrAddr := range GConfig.IndexSvrCluster {
		if svrAddr == localAddr {
			delPos = index
		}
	}
	GConfig.IndexSvrCluster = append(GConfig.IndexSvrCluster[:delPos], GConfig.IndexSvrCluster[delPos+1:]...)

	base.GLog.Debug("IndeSvr Config[%+v]", *GConfig)
	return nil
}
