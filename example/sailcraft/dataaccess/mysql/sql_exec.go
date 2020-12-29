/*
 * @Author: calmwu
 * @Date: 2017-10-18 14:43:29
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-10-18 15:02:44
 * @Comment:
 */

package mysql

import "fmt"

// 直接执行sql命令

func SqlExec(engine *DBEngineInfoS, sqlContent string) (affectedRows, lastInsertId int64, err error) {
	affectedRows = 0
	lastInsertId = -1
	err = nil

	if len(sqlContent) > 0 && engine != nil {
		result, res := engine.RealEngine.Exec(sqlContent)
		if res != nil {
			return
		}
		affectedRows, _ = result.RowsAffected()
		lastInsertId, _ = result.LastInsertId()
	} else {
		err = fmt.Errorf("Params engine or sqlContent is invalid!")
	}

	return
}
