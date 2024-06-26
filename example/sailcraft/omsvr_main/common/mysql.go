/*
 * @Author: calmwu
 * @Date: 2018-05-17 11:13:42
 * @Last Modified by:   calmwu
 * @Last Modified time: 2018-05-17 11:13:42
 * @Comment:
 */

package common

import (
	"fmt"
	"sailcraft/dataaccess/mysql"
	"sailcraft/sysconf"
	"sync"
	"time"
)

var (
	GDBEngine     *mysql.DBEngineInfoS = nil
	mysqlInitOnce sync.Once
)

func InitMysql(mysqlAttr *sysconf.MysqlAttr) error {
	var err error

	mysqlInitOnce.Do(func() {
		dbAddr := fmt.Sprintf("%s:%d", mysqlAttr.Host, mysqlAttr.Port)
		GDBEngine, err = mysql.CreateDBEngnine("mysql", mysqlAttr.User, mysqlAttr.Password, dbAddr, mysqlAttr.Database)

		if err == nil {
			mysql.DoDBKeepAlive(GDBEngine, 300*time.Second)
			mysql.SetDBEngineConnectionParams(GDBEngine, 20, 5)
		}
	})
	return err
}
