package mysql

import (
	utils "calmwu-go/utils"
	"errors"
	"fmt"
	"sailcraft/sysconf"
	"time"
)

type MysqlManager struct {
	EngineMap map[string]*DBEngineInfoS
}

const (
	MAX_DB_CONNECTION_COUNT = 20
	MYSQL_PING              = 300
)

var (
	GMysqlManager *MysqlManager = new(MysqlManager)
)

func (mysqlMgr *MysqlManager) GetMysql(dbName string) (*DBEngineInfoS, error) {
	if len(dbName) == 0 {
		return nil, errors.New("parameters dbname is empty")
	}

	if dbEngine, ok := mysqlMgr.EngineMap[dbName]; ok {
		return dbEngine, nil
	} else {
		return nil, errors.New("database is not exist")
	}
}

func (mysqlMgr *MysqlManager) Initialize() error {
	utils.ZLog.Debugf("MysqlManager Initialize enter")

	mysqlMgr.EngineMap = make(map[string]*DBEngineInfoS)

	configMap := sysconf.GMysqlConfig.ConfigMap
	for key, mysqlAttr := range configMap {
		addr := fmt.Sprintf("%s:%d", mysqlAttr.Host, mysqlAttr.Port)
		dbEngine, err := CreateDBEngnine("mysql", mysqlAttr.User, mysqlAttr.Password, addr, mysqlAttr.Database)
		if err != nil {
			utils.ZLog.Errorf("CreateDBEngnine failed reason[%s]", err.Error())
			return err
		}

		SetDBEngineConnectionParams(dbEngine, MAX_DB_CONNECTION_COUNT, MAX_DB_CONNECTION_COUNT)

		DoDBKeepAlive(dbEngine, MYSQL_PING*time.Second)

		mysqlMgr.EngineMap[key] = dbEngine
	}

	return nil
}
