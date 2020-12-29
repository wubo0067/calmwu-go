/*
 * @Author: calmwu
 * @Date: 2018-11-23 19:43:20
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-24 11:44:36
 */

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

type DoyoCountryPushInfo struct {
	Country string `json:"Country"`
	Url     string `json:"Url"`
}

type DoyoPushUrlConfig struct {
	TestEnv   []DoyoCountryPushInfo `json:"Test"`
	OnlineEnv []DoyoCountryPushInfo `json:"Product"`
}

type DoyoPushConfig struct {
	PushUrlConfig DoyoPushUrlConfig `json:"PushUrlConfig"`
}

var (
	pushConfig  *DoyoPushConfig = nil
	configGuard                 = &sync.Mutex{}
)

func loadConfig(configFile string) error {
	configGuard.Lock()
	defer configGuard.Unlock()

	pushConfig = new(DoyoPushConfig)

	confFile, err := os.Open(configFile)
	if err != nil {
		logger.Printf("open [%s] failed! reason[%s]\n", configFile, err.Error())
		return err
	}
	defer confFile.Close()

	confData, err := ioutil.ReadAll(confFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(confData, pushConfig)
	if err != nil {
		logger.Printf("unmarshal [%s] file failed! reason[%s]\n", configFile, err.Error())
		return err
	}
	logger.Printf("%+v", *pushConfig)

	return nil
}

func findPushUrl(country string, env string) (string, error) {
	configGuard.Lock()
	defer configGuard.Unlock()

	var countryPushList []DoyoCountryPushInfo
	if env == "test" {
		countryPushList = pushConfig.PushUrlConfig.TestEnv
	} else if env == "product" {
		countryPushList = pushConfig.PushUrlConfig.OnlineEnv
	} else {
		return "", fmt.Errorf("country[%s] env[%s] is invalid", country, env)
	}

	var defaultPushUrl string
	for i := range countryPushList {
		countryPushInfo := &countryPushList[i]
		if countryPushInfo.Country == country {
			logger.Printf("country[%s] env[%s] pushUrl[%s]", country, env, countryPushInfo.Url)
			return countryPushInfo.Url, nil
		} else {
			if countryPushInfo.Country == "Others" {
				defaultPushUrl = countryPushInfo.Url
			}
		}
	}
	logger.Printf("country[%s] env[%s] pushUrl[%s]", country, env, defaultPushUrl)
	return defaultPushUrl, nil
}
