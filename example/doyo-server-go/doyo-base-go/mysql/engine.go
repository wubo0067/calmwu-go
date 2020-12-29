/*
 * @Author: calmwu
 * @Date: 2017-10-17 10:52:44
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-30 17:26:30
 * @Comment:
 */

package mysql

import (
	base "doyo-server-go/doyo-base-go"
	"fmt"
	"reflect"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
)

type DBEngineInfoS struct {
	DBEngineDriverName    string
	DBEngineDataSouceName string
	RealEngine            *xorm.Engine
}

func CreateDBEngnine(driverName string, dbUser string, dbPwd string, dbAddr string, dbName string) (*DBEngineInfoS, error) {
	if len(driverName) == 0 ||
		len(dbUser) == 0 ||
		len(dbPwd) == 0 ||
		len(dbName) == 0 {
		return nil, fmt.Errorf("The input database parameters are incorrect!")
	}

	dbEngineInfo := new(DBEngineInfoS)

	dbEngineInfo.DBEngineDriverName = driverName
	dbEngineInfo.DBEngineDataSouceName = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", dbUser, dbPwd, dbAddr, dbName)
	if len(dbAddr) == 0 {
		dbEngineInfo.DBEngineDataSouceName = fmt.Sprintf("%s:%s@/%s?charset=utf8", dbUser, dbPwd, dbName)
	}

	engine, err := xorm.NewEngine(dbEngineInfo.DBEngineDriverName, dbEngineInfo.DBEngineDataSouceName)
	if err != nil {
		base.ZLog.Errorf("xorm.NewEngine Failed! reason[%s]", err.Error())
		return nil, err
	}
	dbEngineInfo.RealEngine = engine
	dbEngineInfo.RealEngine.SetMapper(core.SameMapper{})

	return dbEngineInfo, nil
}

func DoDBKeepAlive(engine *DBEngineInfoS, pingInterval time.Duration) {
	base.ZLog.Debug("start doDBKeepAlive")

	ticker := time.NewTicker(pingInterval)
	go func() {
		defer ticker.Stop()

		for t := range ticker.C {
			err := engine.RealEngine.Ping()
			if err != nil {
				base.ZLog.Errorf("Ping dbSource[%s] failed! At:%v, reason:[%s]",
					engine.DBEngineDataSouceName, t, err.Error())
			}
		}
	}()
}

func SetDBEngineLog(engine *DBEngineInfoS, logH core.ILogger, level core.LogLevel) {
	if engine != nil {
		engine.RealEngine.Logger().SetLevel(level)
		engine.RealEngine.SetLogger(logH)
	}
}

func SetDBEngineConnectionParams(engine *DBEngineInfoS, maxConnCount int, maxIdelConnCount int) {
	if engine != nil {
		engine.RealEngine.SetMaxOpenConns(maxConnCount)
		engine.RealEngine.SetMaxIdleConns(maxIdelConnCount)
		base.ZLog.Debugf("%s maxConnCount:%d maxIdleConnCount:%d", engine.DBEngineDataSouceName,
			maxConnCount, maxIdelConnCount)
	}
}

func GetDbMetas(engine *DBEngineInfoS) ([]*core.Table, error) {
	if engine != nil {
		return engine.RealEngine.DBMetas()
	}
	base.ZLog.Errorf("engine is nil!")
	return nil, fmt.Errorf("engine is nil")
}

func GetTableColumnNames(table *core.Table) ([]string, error) {
	if table != nil {
		return table.ColumnsSeq(), nil
	}
	base.ZLog.Errorf("table is nil!")
	return nil, fmt.Errorf("table is nil")
}

func CreateTable(engine *DBEngineInfoS, table interface{}) error {
	if engine != nil {
		// 判断对象是否实现了tablename方法
		if _, ok := reflect.Indirect(reflect.ValueOf(table)).Interface().(xorm.TableName); ok {
			return engine.RealEngine.CreateTables(table)
		} else {
			return fmt.Errorf("table type[%s] does not implement xorm.TableName interface",
				reflect.TypeOf(table).Name())
		}
	}
	return fmt.Errorf("engine is nil")
}

func SetTableName(engine *DBEngineInfoS, tableName string) {
	if engine != nil {
		engine.RealEngine.Table(tableName)
	}
}
