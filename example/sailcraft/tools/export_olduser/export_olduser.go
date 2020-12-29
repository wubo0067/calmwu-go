/*
 * @Author: calmwu
 * @Date: 2018-05-08 15:28:33
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-08 18:16:59
 * @Comment:
 */

package main

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/dataaccess/mysql"
	"strconv"

	"github.com/gocql/gocql"
)

const (
	CTB_UserOnline   = "tbl_UserOnline"
	CTB_UserRecharge = "tbl_UserRecharge"

	AverageOnlineTime = 180 // 平均在线时长180秒，每天
	DAYSCond1_        = 30  // 30天，平均
	DAYSCond2         = 180 // 180天
	CassandraKeySpace = "ks_statisticmodule"
)

type UserInfoS struct {
	Uin       int
	TotalCost float32
	LoginDays int32
	DeviceID  string
}

var (
	// 3个条件的uin
	// cond1UserList = singlylinkedlist.New()
	// cond2UserList = singlylinkedlist.New()
	// cond3UserList = singlylinkedlist.New()
	//
	UserMap        = make(map[int]*UserInfoS)
	cassandraHosts = []string{"10.161.118.95", "10.161.118.88"}

	// new cassandra
	cassandraNewHosts = []string{"10.161.118.71", "10.161.118.83"}
)

func initCassandra() *gocql.Session {
	cluster := gocql.NewCluster(cassandraHosts[0:]...)
	cluster.Keyspace = CassandraKeySpace
	cluster.Consistency = gocql.Quorum

	base.GLog.Debug("Now start cassandra cluster create session!")
	session, err := cluster.CreateSession()
	if err != nil {
		base.GLog.Error("Cassandra cluster:%v create session failed! error:[%s]", cassandraHosts,
			err.Error())
		return nil
	} else {
		base.GLog.Info("Cassandra cluster:%v create session successed!", cassandraHosts)
	}
	return session
}

func initImportCassandra() *gocql.Session {
	cluster := gocql.NewCluster(cassandraNewHosts[0:]...)
	cluster.Keyspace = CassandraKeySpace
	cluster.Consistency = gocql.Quorum

	base.GLog.Debug("Now start cassandra cluster create session!")
	session, err := cluster.CreateSession()
	if err != nil {
		base.GLog.Error("Cassandra cluster:%v create session failed! error:[%s]", cassandraHosts,
			err.Error())
		return nil
	} else {
		base.GLog.Info("Cassandra cluster:%v create session successed!", cassandraHosts)
	}
	return session
}

func createEngine() (*mysql.DBEngineInfoS, error) {
	driverName := "mysql"
	dbUser := "root"
	dbPwd := "cdb!root123"
	dbAddr := "10.28.95.153:3306"
	dbName := "sailcraft_platform_set"

	return mysql.CreateDBEngnine(driverName, dbUser, dbPwd, dbAddr, dbName)
}

func toString(i interface{}) string {
	switch i.(type) {
	case []byte:
		return string(i.([]byte))
	case string:
		return i.(string)
	}
	return fmt.Sprintf("%v", i)
}

func toInt(i interface{}) int {
	switch i.(type) {
	case []byte:
		n, _ := strconv.ParseInt(string(i.([]byte)), 10, 64)
		return int(n)
	case int:
		return (i.(int))
	case int64:
		return int(i.(int64))
	}
	return 0
}

func main() {
	base.InitLog("export.log")
	defer base.GLog.Close()

	// 初始化cassandra，连接老库
	cSession := initCassandra()
	if cSession == nil {
		base.GLog.Error("cSession is nil")
		return
	}

	// 初始化连接数据库
	dbEngine, err := createEngine()
	if err != nil {
		base.GLog.Error(err.Error())
		return
	}

	var uin int
	var totalOnlineSeconds int64
	var totalRecharge float32

	// 读取tbl_UserOnline
	cqlContent := "SELECT uin, totalonlinetime FROM tbl_UserOnline"
	base.GLog.Debug("Start: %s", cqlContent)
	loginIter := cSession.Query(cqlContent).PageSize(20).Iter()

	for loginIter.Scan(&uin, &totalOnlineSeconds) {
		userInfo := new(UserInfoS)
		userInfo.Uin = uin
		userInfo.LoginDays = int32(totalOnlineSeconds / AverageOnlineTime)
		UserMap[uin] = userInfo
		//base.GLog.Debug("Uin[%d] info:%v", uin, userInfo)
	}

	loginIter.Close()

	// 读取tbl_UserRecharge
	cqlContent = "SELECT uin, totalcost FROM tbl_UserRecharge"
	base.GLog.Debug("Start: %s", cqlContent)
	rechargeIter := cSession.Query(cqlContent).PageSize(20).Iter()

	for rechargeIter.Scan(&uin, &totalRecharge) {
		userInfo, ok := UserMap[uin]
		if ok {
			userInfo.TotalCost = totalRecharge
		} else {
			base.GLog.Error("uin[%d] is not exist!", uin)
		}
	}

	rechargeIter.Close()

	// 查表device表
	records, err := dbEngine.RealEngine.QueryInterface("select device_id, uin from device")
	if err != nil {
		base.GLog.Error(err.Error())
		return
	}
	base.GLog.Debug("records len:%d", len(records))
	for index, _ := range records {
		device_id := toString(records[index]["device_id"])
		uin := toInt(records[index]["uin"])

		userInfo, ok := UserMap[uin]
		if ok {
			userInfo.DeviceID = device_id
		} else {
			base.GLog.Error("uin[%d] is not exist!", uin)
		}
	}

	importSession := initImportCassandra()
	if importSession == nil {
		base.GLog.Error("importSession is nil")
		return
	}

	cqlInsertFmt := "INSERT INTO tbl_OldUserCompensation(DeviceID, OldUin, CompensationLevel, receivestatus) values('%s', %d, %d, 0)"
	// 判断条件
	for _, userInfo := range UserMap {
		level := 0
		if userInfo.TotalCost < 100.00 && userInfo.LoginDays < 30 {
			level = 1
		} else if userInfo.TotalCost >= 1000.00 || userInfo.LoginDays >= 180 {
			level = 3
		} else {
			level = 2
		}

		cqlInsert := fmt.Sprintf(cqlInsertFmt, userInfo.DeviceID, userInfo.Uin, level)
		base.GLog.Debug(cqlInsert)
		err := importSession.Query(cqlInsert).Exec()
		if err != nil {
			base.GLog.Error("%s err:%s", cqlInsert, err.Error())
			return
		}
	}

	return
}
