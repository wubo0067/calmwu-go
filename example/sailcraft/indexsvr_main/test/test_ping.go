/*
 * @Author: calmwu
 * @Date: 2017-10-18 17:23:17
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-10-18 17:28:46
 * @Comment:
 */

package main

import (
	"fmt"
	"sailcraft/indexsvr_main/db"
	"time"
)

func createEngine() (*db.DBEngineInfoS, error) {
	driverName := "mysql"
	dbUser := "root"
	dbPwd := "root"
	dbAddr := "123.59.40.19:13309"
	dbName := "calmwu"

	return db.CreateDBEngnine(driverName, dbUser, dbPwd, dbAddr, dbName)
}

func main() {
	engine, err := createEngine()
	if err != nil {
		fmt.Printf(err.Error())
	} else {
		fmt.Printf("create DBEngine successed!\n")
	}

	db.DoDBKeepAlive(engine, time.Second)

	time.Sleep(time.Second * 10)

	fmt.Println("-----------")
}
