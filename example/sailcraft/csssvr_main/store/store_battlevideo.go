/*
 * @Author: calmwu
 * @Date: 2018-01-27 15:40:06
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-01-27 16:59:18
 * @Comment:
 */

package store

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/csssvr_main/common"
	"sailcraft/csssvr_main/proto"
	"time"
)

func processSvrUploadBattleVideo(cwm *CassandraWorkerMgr, cpd *proto.CassandraProcDataS) {
	uin := cpd.ReqData.Uin
	currentTime := common.GetCassandraMillionSeconds()

	session := CasMgr.GetSessionByKeyspace("ks_statisticmodule")
	if session == nil {
		return
	}

	params, ok := cpd.ReqData.ReqData.Params.(map[string]interface{})
	if !ok {
		base.GLog.Error("Uin[%d] Convert[cpd.ReqData.ReqData.Params] ===> [map[string]interface{}] failed!",
			uin)
		return
	}

	// 获取参数
	var battleVideoID string
	var videoContent string

	if valI, ok := params["BattleVideoID"]; ok {
		if battleVideoID, ok = valI.(string); !ok {
			base.GLog.Error("Uin[%d] Field:BattleVideoID is not string!", uin)
			return
		}
	} else {
		base.GLog.Error("Uin[%d] Field:BattleVideoID not in params!", uin)
		return
	}

	if valI, ok := params["VideoContent"]; ok {
		if videoContent, ok = valI.(string); !ok {
			base.GLog.Error("Uin[%d] Field:VideoContent is not string!", uin)
			return
		}
	} else {
		base.GLog.Error("Uin[%d] Field:VideoContent not in params!", uin)
		return
	}

	var casBattleVideoID string
	var casVideoContent string
	var casRefCount int
	var casTime int64
	applied, err := session.Query(`INSERT INTO tbl_battlevideo(battlevideoid, time, refcount, content) VALUES(?, ?, ?, ?) IF NOT EXISTS`,
		battleVideoID, currentTime, 1, []byte(videoContent)).ScanCAS(&casBattleVideoID, &casVideoContent, &casRefCount, &casTime)
	if err != nil {
		base.GLog.Error("Uin[%d] battleVideoID[%s] Insert into tbl_battlevideo failed! reason[%s]", uin, battleVideoID, err.Error())
		return
	} else {
		if !applied {
			tryUpdateRefCountTimes := 0
			for tryUpdateRefCountTimes < 10 {
				// 录像已经存在，修改引用计数
				refcount := casRefCount + 1
				cqlUpdateBattleVideoRefCount := fmt.Sprintf("UPDATE tbl_battlevideo SET refcount=%d WHERE battlevideoid='%s' IF refcount=%d",
					refcount, battleVideoID, casRefCount)
				applied, err = session.Query(cqlUpdateBattleVideoRefCount).ScanCAS(&casRefCount)
				if err != nil {
					base.GLog.Error("Uin[%d] battleVideoID[%s] exec cql[%s] failed! reason[%s]", uin, battleVideoID, cqlUpdateBattleVideoRefCount, err.Error())
					return
				} else {
					if applied {
						base.GLog.Debug("Uin[%d] battleVideoID[%s] exec cql[%s] successed!", uin, battleVideoID, cqlUpdateBattleVideoRefCount)
						return
					} else {
						base.GLog.Warn("Uin[%d] battleVideoID[%s] exec cql[%s] not applied! try do, count[%d]", uin, battleVideoID, cqlUpdateBattleVideoRefCount, tryUpdateRefCountTimes)
						time.Sleep(time.Second)
					}
				}
				tryUpdateRefCountTimes++
			}
		} else {
			base.GLog.Debug("Uin[%d] battleVideoID[%s] insert into tbl_battlevideo successed!", uin, battleVideoID)
		}
	}
}

func processSvrDeleteBattleVideo(cwm *CassandraWorkerMgr, cpd *proto.CassandraProcDataS) {
	uin := cpd.ReqData.Uin

	session := CasMgr.GetSessionByKeyspace("ks_statisticmodule")
	if session == nil {
		return
	}

	params, ok := cpd.ReqData.ReqData.Params.(map[string]interface{})
	if !ok {
		base.GLog.Error("Uin[%d] Convert[cpd.ReqData.ReqData.Params] ===> [map[string]interface{}] failed!",
			uin)
		return
	}

	var battleVideoID string

	if valI, ok := params["BattleVideoID"]; ok {
		if battleVideoID, ok = valI.(string); !ok {
			base.GLog.Error("Uin[%d] Field:BattleVideoID is not string!", uin)
			return
		}
	} else {
		base.GLog.Error("Uin[%d] Field:BattleVideoID not in params!", uin)
		return
	}

	// 在满足refcount=1的情况下删除录像
	var casRefCount int
	cqlDelBattleVideo := fmt.Sprintf("DELETE FROM tbl_battlevideo where battlevideoid='%s' IF refcount=1", battleVideoID)
	applied, err := session.Query(cqlDelBattleVideo).ScanCAS(&casRefCount)
	if err != nil {
		base.GLog.Error("Uin[%d] exec cql[%s] failed! reason[%s]", uin, cqlDelBattleVideo, err.Error())
		return
	} else {
		if !applied && casRefCount > 0 {
			// 减少引用计数
			tryUpdateRefCountTimes := 0
			for tryUpdateRefCountTimes < 3 {
				refCount := casRefCount - 1
				base.GLog.Debug("Uin[%d] battleVideoID[%s] casRefCount[%d] refCount[%d]", uin, battleVideoID, casRefCount, refCount)

				cqlUpdateBattleVideoRefCount := fmt.Sprintf("UPDATE tbl_battlevideo SET refcount=%d WHERE battlevideoid='%s' IF refcount=%d",
					refCount, battleVideoID, casRefCount)
				applied, err = session.Query(cqlUpdateBattleVideoRefCount).ScanCAS(&casRefCount)
				if err != nil {
					base.GLog.Error("Uin[%d] exec cql[%s] failed! reason[%s]", uin, cqlUpdateBattleVideoRefCount, err.Error())
					return
				} else {
					if applied {
						base.GLog.Debug("Uin[%d] exec cql[%s] successed!", uin, cqlUpdateBattleVideoRefCount)
						return
					} else {
						base.GLog.Warn("Uin[%d] exec cql[%s] not applied! try do, count[%d]", uin, cqlUpdateBattleVideoRefCount, tryUpdateRefCountTimes)
						time.Sleep(100 * time.Millisecond)

						if casRefCount == 1 {
							// 如果已经是等于1了，直接删除
							cqlDelBattleVideo = fmt.Sprintf("DELETE FROM tbl_battlevideo where battlevideoid='%s'", battleVideoID)
							execCql(session, cqlDelBattleVideo)
							base.GLog.Debug("Uin[%d] battleVideoID[%s] refcount=1 so delete this record!", uin, battleVideoID)
							return
						}
					}
				}
				tryUpdateRefCountTimes++
			}
		} else {
			base.GLog.Debug("Uin[%d] exec cql[%s] successed!", uin, cqlDelBattleVideo)
		}
	}
}

func processGetBattleVideo(cwm *CassandraWorkerMgr, cpd *proto.CassandraProcDataS) {
	cassandraProcResult := &proto.CassandraProcResultS{
		Ok:     false,
		Result: nil,
	}

	uin := cpd.ReqData.Uin
	session := CasMgr.GetSessionByKeyspace("ks_statisticmodule")
	if session == nil {
		return
	}

	params, ok := cpd.ReqData.ReqData.Params.(map[string]interface{})
	if !ok {
		base.GLog.Error("Convert[cpd.ReqData.ReqData.Params] ===> [map[string]interface{}] failed!")
		return
	}

	var battleVideoID string
	if valI, ok := params["BattleVideoID"]; ok {
		if battleVideoID, ok = valI.(string); !ok {
			base.GLog.Error("Uin[%d] Field:BattleVideoID is not string!", uin)
			return
		}
	} else {
		base.GLog.Error("Uin[%d] Field:BattleVideoID not in params!", uin)
		return
	}

	cqlQueryBattleVideo := fmt.Sprintf("SELECT content FROM tbl_battlevideo WHERE battlevideoid='%s'", battleVideoID)
	queryRes := QueryRecords(session, cqlQueryBattleVideo)
	if queryRes != nil && len(queryRes) == 1 {
		if valI, ok := queryRes[0]["content"]; ok {
			if battleVideoContent, ok := valI.([]byte); ok {
				var battleVideoRes proto.ProtoGetBattleVideoParamsResS
				battleVideoRes.BattleVideoID = battleVideoID
				battleVideoRes.VideoContent = string(battleVideoContent)
				cassandraProcResult.Ok = true
				cassandraProcResult.Result = &battleVideoRes
			} else {
				base.GLog.Error("Uin[%d] content is not string!", uin)
			}
		} else {
			base.GLog.Error("Uin[%d] query record[%+v] has not content field!", uin, queryRes[0])
		}
	} else {
		base.GLog.Error("Uin[%d] exec cql[%s] failed!", uin, cqlQueryBattleVideo)
	}

	cpd.ResultChan <- cassandraProcResult
	base.GLog.Debug("Uin[%d] query battlevideo[%s] return!", uin, battleVideoID)
}
