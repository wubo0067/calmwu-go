package sysconf

import "fmt"

const (
	MYSQL_CONFIG_NAME = "mysql_config.json"
	REDIS_CONFIG_NAME = "redis_config.json"
)

func Initialize(path string) error {
	mysqlFilePath := fmt.Sprintf("%s/%s", path, MYSQL_CONFIG_NAME)
	err := GMysqlConfig.Init(mysqlFilePath)
	if err != nil {
		return err
	}

	redisFilePath := fmt.Sprintf("%s/%s", path, REDIS_CONFIG_NAME)
	err = GRedisConfig.Init(redisFilePath)
	if err != nil {
		return err
	}

	return nil
}
