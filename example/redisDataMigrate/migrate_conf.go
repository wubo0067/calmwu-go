package main

import "os"
import "io/ioutil"
import "encoding/json"
import "fmt"
import "errors"

type RedisSvrInfo struct {
	RedisSvrAddr string `json:"RedisSvrAddr"`
	RedisSvrAuth string `json:"Auth"`
}

type Config struct {
	MigrateFromServs   []RedisSvrInfo `json:"MigrateFromServs"`
	MigrateToServs     []RedisSvrInfo `json:"MigrateToServs"`
	ScanCount          int            `json:"ScanCount"`
	LScanCount         int            `json:"LScanCount"`
	SScanCount         int            `json:"SScanCount"`
	HScanCount         int            `json:"HScanCount"`
	MigrateWorkerCount int            `json:"MigrateWorkerCount"`
}

func ParseConfig(conf_path string) (*Config, error) {
	conf_file, err := os.Open(conf_path)
	if err != nil {
		fmt.Printf("open [%s] failed! err[%s]\n", conf_path, err.Error())
		return nil, err
	}
	defer conf_file.Close()

	data, err := ioutil.ReadAll(conf_file)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Printf("unmarshal [%s] file failed! err[%s]\n", conf_path, err.Error())
		return nil, err
	}

	gLog.Debug("config data: [%+v]", config)

	if len(config.MigrateFromServs) == 0 ||
		len(config.MigrateToServs) == 0 {
		gLog.Critical("config data is invalid!")
		return nil, errors.New("config data is invalid!")
	}

	return &config, nil
}
