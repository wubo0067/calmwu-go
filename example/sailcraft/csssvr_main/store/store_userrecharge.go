/*
 * @Author: calmwu
 * @Date: 2018-01-29 19:29:06
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-16 18:22:00
 * @Comment:
 */
package store

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/csssvr_main/common"
	"sailcraft/csssvr_main/proto"
	"sync"
	"time"

	"github.com/emirpasic/gods/sets/hashset"
	"github.com/mitchellh/mapstructure"
)

var (
	initRevenueOnce            sync.Once
	statisticsGameRechargeChan chan *proto.ProtoCssSvrUserRechargeNtf
)

func processUserRecharge(cwm *CassandraWorkerMgr, cpd *proto.CassandraProcDataS) {
	uin := cpd.ReqData.Uin

	initRevenueOnce.Do(func() {
		statisticsGameRechargeChan = make(chan *proto.ProtoCssSvrUserRechargeNtf, 10240)

		go func() {
			for userRechargeNtf := range statisticsGameRechargeChan {
				statisticsGameDailyRevenue(userRechargeNtf)
				statisticsCountryDailyRevenue(userRechargeNtf)
			}
		}()
	})

	session := CasMgr.GetSessionByKeyspace("ks_statisticmodule")
	if session == nil {
		return
	}

	var reqData proto.ProtoCssSvrUserRechargeNtf
	err := mapstructure.Decode(cpd.ReqData.ReqData.Params, &reqData)
	if err != nil {
		base.GLog.Error("Uin[%d] Decode ProtoCssSvrUserRechargeNtf failed! reason[%s]",
			uin, err.Error())
		return
	}

	statisticsGameRechargeChan <- &reqData

	keys := hashset.New()
	keys.Add("Uin")
	keys.Add("ChannelID")

	// 查询更新tbl_UserTotalRecharge, 单个用户并发充值不可能，不必使用cas
	cqlSelectUserRecharge := fmt.Sprintf("SELECT * FROM tbl_UserTotalRecharge WHERE uin=%d", uin)
	userTotalRechargeRecords := QueryRecords(session, cqlSelectUserRecharge)
	if userTotalRechargeRecords != nil {
		if len(userTotalRechargeRecords) > 0 {
			var userTotalRecharge TblUserTotalRechargeS
			err := mapstructure.Decode(userTotalRechargeRecords[0], &userTotalRecharge)
			if err == nil {
				userTotalRecharge.RechargeCount++
				userTotalRecharge.TotalCost += reqData.RechargeAmount
				cqlUpdateUserRecharge, _ := genUpdateCql(TBNAME_USERRECHARGE, &userTotalRecharge, keys)
				execCql(session, cqlUpdateUserRecharge)
			} else {
				base.GLog.Error("Uin[%d] Decode TblUserTotalRechargeS object failed! reason[%s]",
					err.Error())
				return
			}
		} else {
			// 插入记录
			cqlInsertUserTotalRecharge := fmt.Sprintf("INSERT INTO tbl_UserTotalRecharge(Uin, ChannelID, TotalCost, RechargeCount, Platform) VALUES(%d, '%s', %f, 1, '%s')",
				uin, reqData.ChannelID, reqData.RechargeAmount, reqData.PlatForm)
			execCql(session, cqlInsertUserTotalRecharge)
		}
	}

	// 根据uin查询玩家的国家iso
	var isoCountryCode string
	if err := session.Query("SELECT ISOCountryCode FROM tbl_UserOnline where uin = ?", uin).Scan(&isoCountryCode); err != nil {
		base.GLog.Error("Uin[%d] query field[ISOCountryCode] failed! reason[%s]", uin, err.Error())
		return
	}

	currentTime := common.GetCassandraMillionSeconds()

	// tbl_UserRechargeRecord  充值详细记录
	cqlInsertUserRechargeRecord := fmt.Sprintf("INSERT INTO tbl_UserRechargeRecord(Uin, ChannelID, ISOCountryCode, Platform, Cost, Time) VALUES(%d, '%s', '%s', '%s', %f, %d)",
		uin, reqData.ChannelID, isoCountryCode, reqData.PlatForm, reqData.RechargeAmount, currentTime)
	execCql(session, cqlInsertUserRechargeRecord)

	return
}

func statisticsGameDailyRevenue(req *proto.ProtoCssSvrUserRechargeNtf) {
	session := CasMgr.GetSessionByKeyspace("ks_statisticmodule")
	if session == nil {
		return
	}

	date := base.GetDate()

	// 更新日收入表，这里要用到cas了，保证版本号大于基础的，因为多个服务实例
	cqlInsertDailyRevenue := fmt.Sprintf("INSERT INTO tbl_DailyRevenue(Date, ChannelID, Platform, TotalRevenue, TotalRechargeCount, VersionID) VALUES('%s', '%s', '%s', %f, 1, 1) IF NOT EXISTS",
		date, req.ChannelID, req.PlatForm, req.RechargeAmount)

	var casDate string
	var casChannelID string
	var casPlatform string
	var casTotalRevenue float32
	var casTotalRechargeCount int
	var casVersionID int

	applied, err := session.Query(cqlInsertDailyRevenue).ScanCAS(&casDate, &casChannelID, &casPlatform,
		&casTotalRechargeCount, &casTotalRevenue, &casVersionID)
	if err != nil {
		base.GLog.Error("exec cql[%s] failed! reason[%s]", cqlInsertDailyRevenue, err.Error())
		return
	} else if !applied {
		// 数据存在，需要更新操作
		// 日，json中默认是float64，cassandra中默认是float32，操蛋
		base.GLog.Warn("exec[%s] not applied! casTotalRevenue[%f] casTotalRechargeCount[%d] casVersionId[%d]",
			cqlInsertDailyRevenue, casTotalRevenue, casTotalRechargeCount, casVersionID)

		totalRevenue := casTotalRevenue + req.RechargeAmount
		totalRechargeCount := casTotalRechargeCount + 1
		versionID := casVersionID + 1

		cqlUpdateDailyRevenueFmt := "UPDATE tbl_DailyRevenue SET TotalRevenue=%f, TotalRechargeCount=%d, VersionID=%d WHERE Date='%s' AND ChannelID='%s' AND Platform='%s' IF TotalRevenue=%f and TotalRechargeCount=%d and VersionID=%d"
		var tryCount = 10
		for tryCount > 0 {
			cqlUpdateDailyRevenue := fmt.Sprintf(cqlUpdateDailyRevenueFmt,
				totalRevenue, totalRechargeCount, versionID, date,
				casChannelID, casPlatform,
				casTotalRevenue, casTotalRechargeCount, casVersionID)
			// 这种模式其实在竞争很多的时候效果不好
			applied, err = session.Query(cqlUpdateDailyRevenue).ScanCAS(&casTotalRevenue,
				&casTotalRechargeCount, &casVersionID)
			if err != nil {
				base.GLog.Error("exec cql[%s] failed! reason[%s]", cqlUpdateDailyRevenue, err.Error())
				break
			} else if !applied {
				base.GLog.Warn("Not Applied! tryCount[%d] exec cql[%s] casTotalRevenue[%f] casTotalRechargeCount[%d] casVersionId[%d]",
					tryCount, cqlUpdateDailyRevenue, casTotalRevenue, casTotalRechargeCount, casVersionID)
				// 版本号冲突，需要递增
				versionID = casVersionID + 1
				totalRevenue = casTotalRevenue + req.RechargeAmount
				totalRechargeCount = casTotalRechargeCount + 1
				time.Sleep(10 * time.Millisecond)
				//runtime.Gosched()
			} else {
				base.GLog.Debug("exec[%s] successed!", cqlUpdateDailyRevenue)
				break
			}
			tryCount--
			//common.GLog.Debug("tryCount[%d]", tryCount)
		}

		if tryCount <= 0 {
			base.GLog.Error("++++UPDATE tbl_DailyRevenue++++ failed!")
		}
	} else {
		base.GLog.Debug("exec[%s] successed!", cqlInsertDailyRevenue)
	}
}

func statisticsCountryDailyRevenue(req *proto.ProtoCssSvrUserRechargeNtf) {
	uin := req.Uin
	date := base.GetDate()

	session := CasMgr.GetSessionByKeyspace("ks_statisticmodule")
	if session == nil {
		return
	}

	var isoCountryCode string
	if err := session.Query("SELECT ISOCountryCode FROM tbl_UserOnline where uin = ?", uin).Scan(&isoCountryCode); err != nil {
		base.GLog.Error("Uin[%d] query field[ISOCountryCode] failed! reason[%s]", uin, err.Error())
		return
	}

	// 更新日收入表，这里要用到cas了，保证版本号大于基础的，因为多个服务实例
	cqlInsertDailyCountryRevenue := fmt.Sprintf("INSERT INTO tbl_DailyCountryRevenue(Date, ISOCountryCode, ChannelID, Platform, TotalRevenue, TotalRechargeCount, VersionID) VALUES('%s', '%s', '%s', '%s', %f, 1, 1) IF NOT EXISTS",
		date, isoCountryCode, req.ChannelID, req.PlatForm, req.RechargeAmount)

	var casDate string
	var casIsoCountryCode string
	var casChannelID string
	var casPlatform string
	var casTotalRevenue float32
	var casTotalRechargeCount int
	var casVersionID int

	// 这里一定要注意实际的表中顺序
	applied, err := session.Query(cqlInsertDailyCountryRevenue).ScanCAS(&casDate, &casIsoCountryCode, &casChannelID, &casPlatform,
		&casTotalRechargeCount, &casTotalRevenue, &casVersionID)
	if err != nil {
		base.GLog.Error("exec cql[%s] failed! reason[%s]", cqlInsertDailyCountryRevenue, err.Error())
		return
	} else if !applied {
		// 数据存在，需要更新操作
		// 日，json中默认是float64，cassandra中默认是float32，操蛋
		base.GLog.Warn("exec[%s] not applied! casTotalRevenue[%f] casTotalRechargeCount[%d] casVersionId[%d]",
			cqlInsertDailyCountryRevenue, casTotalRevenue, casTotalRechargeCount, casVersionID)

		totalRevenue := casTotalRevenue + req.RechargeAmount
		totalRechargeCount := casTotalRechargeCount + 1
		versionID := casVersionID + 1

		cqlUpdateDailyCountryRevenueFmt := "UPDATE tbl_DailyCountryRevenue SET TotalRevenue=%f, TotalRechargeCount=%d, VersionID=%d WHERE Date='%s' AND ISOCountryCode='%s' AND ChannelID='%s' AND Platform='%s' IF TotalRevenue=%f and TotalRechargeCount=%d and VersionID=%d"
		var tryCount = 10
		for tryCount > 0 {
			cqlUpdateDailyCountryRevenue := fmt.Sprintf(cqlUpdateDailyCountryRevenueFmt,
				totalRevenue, totalRechargeCount, versionID, date,
				casIsoCountryCode, casChannelID, casPlatform,
				casTotalRevenue, casTotalRechargeCount, casVersionID)
			// 这种模式其实在竞争很多的时候效果不好
			applied, err = session.Query(cqlUpdateDailyCountryRevenue).ScanCAS(&casTotalRevenue,
				&casTotalRechargeCount, &casVersionID)
			if err != nil {
				base.GLog.Error("exec cql[%s] failed! reason[%s]", cqlUpdateDailyCountryRevenue, err.Error())
				break
			} else if !applied {
				base.GLog.Warn("Not Applied! tryCount[%d] exec cql[%s] casTotalRevenue[%f] casTotalRechargeCount[%d] casVersionId[%d]",
					tryCount, cqlUpdateDailyCountryRevenue, casTotalRevenue, casTotalRechargeCount, casVersionID)
				// 版本号冲突，需要递增
				versionID = casVersionID + 1
				totalRevenue = casTotalRevenue + req.RechargeAmount
				totalRechargeCount = casTotalRechargeCount + 1
				time.Sleep(10 * time.Millisecond)
				//runtime.Gosched()
			} else {
				base.GLog.Debug("exec[%s] successed!", cqlUpdateDailyCountryRevenue)
				break
			}
			tryCount--
			//common.GLog.Debug("tryCount[%d]", tryCount)
		}

		if tryCount <= 0 {
			base.GLog.Error("++++UPDATE tbl_DailyRevenue++++ failed!")
		}
	} else {
		base.GLog.Debug("exec[%s] successed!", cqlInsertDailyCountryRevenue)
	}
}

func processGetUserRechargeInfo(cwm *CassandraWorkerMgr, cpd *proto.CassandraProcDataS) {
	cassandraProcResult := &proto.CassandraProcResultS{
		Ok:     false,
		Result: nil,
	}

	uin := cpd.ReqData.Uin
	session := CasMgr.GetSessionByKeyspace("ks_statisticmodule")
	if session == nil {
		return
	}

	// var reqData proto.ProtoQueryUserRechargeInfoReq
	// err := mapstructure.Decode(cpd.ReqData.ReqData.Params, &reqData)
	// if err != nil {
	// 	base.GLog.Error("Uin[%d] Decode ProtoQueryUserRechargeInfoReq failed! reason[%s]",
	// 		uin, err.Error())
	// 	return
	// }
	var userRechargeInfoRes proto.ProtoQueryUserRechargeInfoRes
	userRechargeInfoRes.Uin = uin
	cassandraProcResult.Result = &userRechargeInfoRes

	cqlQueryUserRechargeRecords := fmt.Sprintf("SELECT * FROM tbl_UserRechargeRecord WHERE uin=%d", uin)
	userRechargeRecords := QueryRecords(session, cqlQueryUserRechargeRecords)
	if userRechargeRecords != nil && len(userRechargeRecords) >= 1 {
		userRechargeInfoRes.TotalRechargeCount = len(userRechargeRecords)
		for index := range userRechargeRecords {
			rechargeRecord := userRechargeRecords[index]

			cost := rechargeRecord["cost"].(float32)
			if cost > userRechargeInfoRes.MaxRechargeAmount {
				userRechargeInfoRes.MaxRechargeAmount = cost
			}
			userRechargeInfoRes.TotalRechargeAmount += cost
		}
	}

	cpd.ResultChan <- cassandraProcResult
	base.GLog.Debug("Uin[%d] query recharge info:%+v return!", uin, userRechargeInfoRes)
}
