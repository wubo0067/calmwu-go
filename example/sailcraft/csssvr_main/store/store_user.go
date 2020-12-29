/*
 * @Author: calmwu
 * @Date: 2018-01-29 19:23:44
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-01-29 19:29:39
 * @Comment:
 */

package store

import (
	"fmt"
	"net"
	"sailcraft/base"
	"sailcraft/csssvr_main/common"
	"sailcraft/csssvr_main/proto"

	"github.com/emirpasic/gods/sets/hashset"
	"github.com/gocql/gocql"
	"github.com/mitchellh/mapstructure"
)

func processUserLogin(cwm *CassandraWorkerMgr, cpd *proto.CassandraProcDataS) {
	uin := cpd.ReqData.Uin
	keys := hashset.New()
	date := base.GetDate()
	currentTime := common.GetCassandraMillionSeconds()

	session := CasMgr.GetSessionByKeyspace("ks_statisticmodule")
	if session == nil {
		return
	}

	var reqData proto.ProtoCssSvrUserLoginNtf
	err := mapstructure.Decode(cpd.ReqData.ReqData.Params, &reqData)
	if err != nil {
		base.GLog.Error("Uin[%d] Decode ProtoCssSvrUserLoginNtf failed! reason[%s]",
			uin, err.Error())
		return
	}

	// 插入日活跃uin表
	updateGameDau(session, uin, reqData.ClientInternetIP, reqData.Platform)

	// 查询tbl_UserOnline
	cqlSelectUserInfo := fmt.Sprintf("SELECT * FROM tbl_UserOnline WHERE uin=%d", uin)
	userInfoRes := QueryRecords(session, cqlSelectUserInfo)

	var userOnline TblUserOnineS
	if userInfoRes != nil {
		keys.Add("Uin")
		if len(userInfoRes) == 0 {
			// 用户不存在，插入用户信息表
			userOnline.Uin = uin
			userOnline.CreateTime = currentTime
			userOnline.ISOCountryCode, _ = common.QueryGeoInfo(reqData.ClientInternetIP)
			userOnline.LoginTime = currentTime
			userOnline.MaxOnlinetime = 0
			userOnline.TotalOnlineTime = 0
			userOnline.VersionID = 1
			userOnline.Platform = reqData.Platform
			cqlInsertUserInfo, _ := genUpdateCql(TBNAME_USERONLINE, &userOnline, keys)
			execCql(session, cqlInsertUserInfo)
			// 日注册
			cqlUpdateDailyRegister := fmt.Sprintf("UPDATE tbl_DailyRegisterCount set registercount=registercount+1 WHERE date='%s'", date)
			execCql(session, cqlUpdateDailyRegister)
			// 插入日注册uin表
			cqlInsertDataRegisterUin := fmt.Sprintf("INSERT INTO tbl_dateregisteruin_%s(uin) VALUES(%d);", date, uin)
			execCql(session, cqlInsertDataRegisterUin)
			// 国家注册
			date := base.GetDate()
			isoCountryName, country := common.QueryGeoInfo(reqData.ClientInternetIP)
			base.GLog.Debug("clientInternetIP[%s] IsoCountryName[%s] country[%s]", reqData.ClientInternetIP, isoCountryName, country)
			cqlUpdateDailyCountryRegister := fmt.Sprintf("UPDATE tbl_DailyCountryRegisterCount set registercount=registercount+1 WHERE date='%s' AND isocountrycode='%s' AND countryname='%s'",
				date, isoCountryName, country)
			execCql(session, cqlUpdateDailyCountryRegister)
			// 平台注册
			cqlUpdateDailyPlatformRegister := fmt.Sprintf("UPDATE tbl_DailyPlatformRegisterCount set RegisterCount=RegisterCount+1 WHERE date='%s' AND Platform='%s'",
				date, reqData.Platform)
			execCql(session, cqlUpdateDailyPlatformRegister)
		} else {
			//
			err := decodeTblUserOnlineRecord(userInfoRes[0], &userOnline)
			if err != nil {
				base.GLog.Error("Decode TblUserOnineS object failed! reason[%s]", err.Error())
				return
			}
			base.GLog.Debug("%+v", userOnline)
			userOnline.LoginTime = currentTime
			userOnline.ISOCountryCode, _ = common.QueryGeoInfo(reqData.ClientInternetIP)
			userOnline.VersionID++
			cqlUpdateUserOnline, _ := genUpdateCql(TBNAME_USERONLINE, &userOnline, keys)
			execCql(session, cqlUpdateUserOnline)
		}
	}

	return
}

func processUserLogout(cwm *CassandraWorkerMgr, cpd *proto.CassandraProcDataS) {
	uin := cpd.ReqData.Uin
	currentTime := common.GetCassandraMillionSeconds()

	session := CasMgr.GetSessionByKeyspace("ks_statisticmodule")
	cqlSelectUserOnline := fmt.Sprintf("SELECT * FROM %s WHERE uin=%d", TBNAME_USERONLINE, uin)
	userOnlineRes := QueryRecords(session, cqlSelectUserOnline)
	if userOnlineRes != nil && len(userOnlineRes) > 0 {
		var userOnline TblUserOnineS
		err := decodeTblUserOnlineRecord(userOnlineRes[0], &userOnline)
		if err != nil {
			base.GLog.Error("Decode TblUserOnineS object failed! reason[%s]", err.Error())
			return
		}
		//keys.Add("Uin")
		onlineSeconds := (currentTime - userOnline.LoginTime) / 1000
		userOnline.LogoutTime = currentTime
		// 总共在线时长
		userOnline.TotalOnlineTime += onlineSeconds
		if onlineSeconds > userOnline.MaxOnlinetime {
			// 最大在线时长
			userOnline.MaxOnlinetime = onlineSeconds
		}
		keys := hashset.New()
		keys.Add("Uin")
		cqlUpdateUserOnline, _ := genUpdateCql(TBNAME_USERONLINE, &userOnline, keys)
		execCql(session, cqlUpdateUserOnline)
	}
}

func updateGameDau(session *gocql.Session, uin int, clientInternetIP string, platform string) {
	// 插入日活跃uin表
	date := base.GetDate()
	cqlInsertDateActiveUin := fmt.Sprintf("INSERT INTO tbl_dateactiveuin_%s(uin) VALUES(%d) IF NOT EXISTS", date, uin)

	var recordUin int
	applied, err := session.Query(cqlInsertDateActiveUin).ScanCAS(&recordUin)
	if err != nil {
		base.GLog.Error("exec cql[%s] failed! reason[%s]", cqlInsertDateActiveUin, err.Error())
	} else {
		if applied {
			// 今天首次登陆
			base.GLog.Debug("+++++Uin[%d] is dau player!", uin)

			// 更新游戏dau
			cqlUpdateDau := fmt.Sprintf("UPDATE tbl_dau set logincount=logincount+1 WHERE date='%s'", date)
			execCql(session, cqlUpdateDau)

			isoCountryName, country := common.QueryGeoInfo(clientInternetIP)
			cqlUpdateCountryDau := fmt.Sprintf("UPDATE tbl_countrydau set logincount=logincount+1 WHERE date='%s' AND isocountrycode='%s' AND countryname='%s'",
				date, isoCountryName, country)
			// 更新国家dau
			execCql(session, cqlUpdateCountryDau)

			// 更新平台dau
			cqlUpdatePlatformDau := fmt.Sprintf("UPDATE tbl_PlatformDau set LoginCount=LoginCount+1 WHERE date='%s' AND Platform='%s'", date, platform)
			execCql(session, cqlUpdatePlatformDau)
		} else {
			base.GLog.Debug("------Uin[%d] is not dau player!", uin)
		}
	}
	return
}

func processQueryPlayerGeo(cwm *CassandraWorkerMgr, cpd *proto.CassandraProcDataS) {
	cassandraProcResult := &proto.CassandraProcResultS{
		Ok:     false,
		Result: nil,
	}

	uin := cpd.ReqData.Uin
	session := CasMgr.GetSessionByKeyspace("ks_statisticmodule")
	if session == nil {
		return
	}

	var queryGeoParams proto.ProtoSvrQueryISOCountryCodesByUinsParamsReqS
	err := mapstructure.Decode(cpd.ReqData.ReqData.Params, &queryGeoParams)
	if err != nil {
		base.GLog.Error("Uin[%d] Decode ProtoSvrQueryISOCountryCodesByUinsParamsReqS failed! reason[%s]",
			uin, err.Error())
		return
	}

	base.GLog.Debug("Uin[%d] ProtoSvrQueryISOCountryCodesByUinsParamsReqS[%+v]", uin, queryGeoParams)

	if queryGeoParams.Count > MAX_GEOUIN_COUNT {
		base.GLog.Error("Uin[%d] Count[%d] More than the MAX_GEOUIN_COUNT limit", uin, queryGeoParams.Count)
		return
	}

	cqlQueryGeos := fmt.Sprintf("SELECT uin, registerregion FROM tbl_useronline WHERE uin in (%s)", base.ArrayToString(queryGeoParams.Uins, ","))

	queryRes := QueryRecords(session, cqlQueryGeos)
	if queryRes != nil && len(queryRes) > 0 {
		var queryGeoResult proto.ProtoSvrQueryISOCountryCodesByUinsParamsResS
		queryGeoResult.ProtoPlayerGeos = make([]*proto.ProtoPlayerGeoS, 0)

		for index, _ := range queryRes {
			playerGeo := new(proto.ProtoPlayerGeoS)
			err := mapstructure.Decode(queryRes[index], playerGeo)

			IP := net.ParseIP(playerGeo.IsoCountryCode)
			if IP != nil {
				playerGeo.IsoCountryCode = "unknown"
			}

			base.GLog.Debug("index[%d] query record:%v, plyergeo:%v", index, queryRes[index], playerGeo)
			if err != nil {
				base.GLog.Error("Uin[%d] Decode From[%+v] ===> ProtoPlayerGeoS failed! reason[%s]", uin, queryRes[index], err.Error())
			} else {
				queryGeoResult.ProtoPlayerGeos = append(queryGeoResult.ProtoPlayerGeos, playerGeo)
				queryGeoResult.Count++
			}
		}

		cassandraProcResult.Ok = true
		cassandraProcResult.Result = &queryGeoResult

	} else {
		base.GLog.Error("Uin[%d] exec cql[%s] failed!", uin, cqlQueryGeos)
	}
	cpd.ResultChan <- cassandraProcResult
	base.GLog.Debug("Uin[%d] Query player GEO infos Return!", uin)
}
