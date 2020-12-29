/*
 * @Author: calmwu
 * @Date: 2018-03-30 10:28:16
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-06-13 15:02:34
 * @Comment:
 */

package handler

import (
	"encoding/json"
	"fmt"
	"sailcraft/base"
	"sailcraft/base/consul_api"
	csssvr_proto "sailcraft/csssvr_main/proto"
	"sailcraft/financesvr_main/common"
	"sailcraft/financesvr_main/proto"
	"time"

	"github.com/emirpasic/gods/lists/singlylinkedlist"
	"github.com/emirpasic/gods/sets/hashset"
)

func GMConfigSuperGiftActive(req *proto.ProtoGMConfigActiveSuperGiftReq) error {
	for index := range req.SuperGiftConfigs {
		config := &req.SuperGiftConfigs[index]
		err := setActiveInstConfig(req.Uin, req.ZoneID, config, proto.E_ACTIVETYPE_SUPERGIFTPACKAGE, config.Base.ActiveID)
		if err != nil {
			return err
		}
	}
	return nil
}

func GMConfigMissionActive(req *proto.ProtoGMConfigActiveMissionReq) error {
	for index := range req.ActiveMissions {
		config := &req.ActiveMissions[index]
		err := setActiveInstConfig(req.Uin, req.ZoneID, config, proto.E_ACTIVETYPE_MISSION, config.Base.ActiveID)
		if err != nil {
			return err
		}
	}
	return nil
}

func GMConfigExchangeActive(req *proto.ProtoGMConfigActiveExchangeReq) error {
	for index := range req.ActiveExchanges {
		config := &req.ActiveExchanges[index]
		err := setActiveInstConfig(req.Uin, req.ZoneID, config, proto.E_ACTIVETYPE_EXCHANGE, config.Base.ActiveID)
		if err != nil {
			return err
		}
	}
	return nil
}

func GMConfigCDKeyExchangeActive(req *proto.ProtoGMConfigActiveCDKeyExchangeReq) error {
	for index := range req.ActiveCDKeyExchanges {
		config := &req.ActiveCDKeyExchanges[index]
		err := setActiveInstConfig(req.Uin, req.ZoneID, config, proto.E_ACTIVETYPE_CDKEYEXCHANGE, config.Base.ActiveID)
		if err != nil {
			return err
		}
	}
	return nil
}

func setActiveInstConfig(uin uint64, zoneID int32, config interface{}, activeType proto.ActiveType, activeID int) error {
	activeInstConfigKey := fmt.Sprintf(proto.ActiveInstConfigKeyFmt, zoneID, activeType.String(),
		activeID)

	base.GLog.Debug("activeInstConfigKey[%s]", activeInstConfigKey)

	redisData, err := json.Marshal(config)
	if err != nil {
		base.GLog.Error("Uin[%d] ZoneID[%d] Marshal failed! reason[%s]", uin, zoneID, err.Error())
		return err
	}

	// 每个活动实例都有自己的配置
	err = common.GRedis.StringSet(activeInstConfigKey, redisData)
	if err != nil {
		err := fmt.Errorf("Set activeInstConfigKey[%s] data failed! reason[%s]", activeInstConfigKey, err.Error())
		base.GLog.Error(err.Error())
		return err
	}
	base.GLog.Debug("Uin[%d] ZoneID[%d] set activeInstConfigKey[%s] successed!", uin, zoneID, activeInstConfigKey)

	if activeType == proto.E_ACTIVETYPE_CDKEYEXCHANGE {
		cdkeyCountKey := fmt.Sprintf(proto.CDKeyCountKeyFmt, activeInstConfigKey)
		common.GRedis.StringSet(cdkeyCountKey, []byte("0"))
	}
	return nil
}

func getActiveInstBaseConfig(activeType proto.ActiveType, activeInstConfig interface{}) *proto.ActiveBaseConfigInfoS {
	switch activeType {
	case proto.E_ACTIVETYPE_SUPERGIFTPACKAGE:
		return &(activeInstConfig.(*proto.ActiveSuperGiftInfoS).Base)
	case proto.E_ACTIVETYPE_MISSION:
		return &(activeInstConfig.(*proto.ActiveMissionInfoS).Base)
	case proto.E_ACTIVETYPE_EXCHANGE:
		return &(activeInstConfig.(*proto.ActiveExchangeInfoS).Base)
	case proto.E_ACTIVETYPE_CDKEYEXCHANGE:
		return &(activeInstConfig.(*proto.ActiveCDKeyExchangeInfoS).Base)
	}
	return nil
}

// 获取具体活动的配置
func getActiveInstConfig(activeType proto.ActiveType, activeInstConfKey string) (interface{}, error) {
	var activeInstConfig interface{}
	var err error

	if activeType == proto.E_ACTIVETYPE_SUPERGIFTPACKAGE {
		activeInstConfig = new(proto.ActiveSuperGiftInfoS)
	} else if activeType == proto.E_ACTIVETYPE_MISSION {
		activeInstConfig = new(proto.ActiveMissionInfoS)
	} else if activeType == proto.E_ACTIVETYPE_EXCHANGE {
		activeInstConfig = new(proto.ActiveExchangeInfoS)
	} else if activeType == proto.E_ACTIVETYPE_CDKEYEXCHANGE {
		activeInstConfig = new(proto.ActiveCDKeyExchangeInfoS)
	} else {
		err = fmt.Errorf("ActiveType[%d] is invalid!", activeType)
		base.GLog.Error(err)
		return nil, err
	}

	//base.GLog.Debug("activeInstConfKey[%s]", activeInstConfKey)
	err = common.GetStrDataFromRedis(activeInstConfKey, activeInstConfig)
	if err != nil {
		err = fmt.Errorf("Get activeInstConfigKey[%s] failed! reason[%s]",
			activeInstConfKey, err.Error())
		base.GLog.Error(err.Error())
		return nil, err
	}
	//base.GLog.Debug("activeInstConfKey[%s] config:%+v", activeInstConfKey, activeInstConfig)
	return activeInstConfig, nil
}

func CheckActiveConfig(req *proto.ProtoCheckActiveConfigReq) *proto.ProtoCheckActiveConfigRes {
	var res proto.ProtoCheckActiveConfigRes
	res.Uin = req.Uin
	res.ZoneID = req.ZoneID
	res.ActiveIDs = req.ActiveIDs
	res.ActiveType = req.ActiveType
	res.IsExists = make([]int32, len(req.ActiveIDs))

	for index, activeID := range req.ActiveIDs {
		activeInstConfigKey := fmt.Sprintf(proto.ActiveInstConfigKeyFmt, req.ZoneID, req.ActiveType.String(),
			activeID)

		base.GLog.Debug("activeInstConfigKey[%s]", activeInstConfigKey)

		isExist, err := common.GRedis.Exists(activeInstConfigKey)
		if err != nil || !isExist {
			base.GLog.Error(err.Error())
			res.IsExists[index] = 0
		} else {
			res.IsExists[index] = 1
		}
	}

	return &res
}

func getActiveControlConfig(zoneID int32) *proto.RunningActiveMgr {
	ActiveRunningKey := fmt.Sprintf(proto.ActiveRunningKeyFmt, zoneID)
	var runningActiveMgr proto.RunningActiveMgr
	err := common.GetStrDataFromRedis(ActiveRunningKey, &runningActiveMgr)
	if err != nil {
		return nil
	}
	return &runningActiveMgr
}

// 计算具体某个活动的开始结束时间
func calcActiveTime(startTime int64, durationSecs int64) (*time.Time, *time.Time) {
	activeStartTime := time.Unix(startTime, 0)
	activeEndTime := activeStartTime.Add(time.Duration(durationSecs) * time.Second)
	base.GLog.Debug("activeStartTime[%s] activeEndTime[%s]", base.TimeName(activeStartTime), base.TimeName(activeEndTime))
	return &activeStartTime, &activeEndTime
}

func OpenActive(req *proto.ProtoOpenActiveReq) *proto.ProtoControlActiveRes {
	var res proto.ProtoControlActiveRes
	res.Uin = req.Uin
	res.ZoneID = req.ZoneID
	res.ActiveIDs = make([]int, len(req.ActiveControlConfigs))
	res.ControlResults = make([]int32, len(req.ActiveControlConfigs))

	runningActiveMgr := getActiveControlConfig(req.ZoneID)
	if runningActiveMgr == nil {
		runningActiveMgr = proto.CreateRunningActiveMgr(req.ZoneID, common.GRedis)
	}

	for index := range req.ActiveControlConfigs {
		activeCtrl := &req.ActiveControlConfigs[index]

		startTime := activeCtrl.BeginTime()
		endTime := activeCtrl.EndTime()
		base.GLog.Debug("ZoneID[%d] ActiveType[%d:%s] ActiveID[%d] startTime[%s] endTime[%s]",
			req.ZoneID, activeCtrl.ActiveType, activeCtrl.ActiveType.String(), activeCtrl.ActiveID,
			base.TimeName(startTime), base.TimeName(endTime))

		res.ActiveIDs[index] = activeCtrl.ActiveID
		err := runningActiveMgr.OpenActive(activeCtrl)
		if err != nil {
			res.ControlResults[index] = 0
		} else {
			res.ControlResults[index] = 1
		}
	}

	runningActiveMgr.SyncRedis(req.ZoneID, common.GRedis)
	return &res
}

func CloseActive(req *proto.ProtoCloseActiveReq) *proto.ProtoControlActiveRes {
	var res proto.ProtoControlActiveRes
	res.Uin = req.Uin
	res.ZoneID = req.ZoneID
	res.ActiveIDs = req.ActiveIDs
	res.ControlResults = make([]int32, len(req.ActiveIDs))

	runningActiveMgr, err := getRunningActives(req.ZoneID)
	if err != nil {
		return &res
	}

	for index, activeID := range req.ActiveIDs {
		err := runningActiveMgr.CloseActive(req.ActiveType, activeID)
		if err != nil {
			res.ControlResults[index] = 0
		} else {
			res.ControlResults[index] = 1
		}
	}

	runningActiveMgr.SyncRedis(req.ZoneID, common.GRedis)
	return &res
}

// 获取玩家的活动信息，根据活动类型
func GetPlayerActive(req *proto.ProtoGetPlayerActiveReq) (*proto.ProtoGetPlayerActiveRes, error) {
	var res proto.ProtoGetPlayerActiveRes
	var err error

	// 开放的活动列表
	runningActiveMgr, err := getRunningActives(req.ZoneID)
	if err != nil {
		return nil, err
	}

	// 用服务器时间
	now := time.Now()

	// 查询玩家对应的活动信息，没有就插入记录
	playerActiveInfos := make([]*proto.TblPlayerActiveInfo, 0)

	// 过滤出来的活动实例配置信息，key是rediskey
	activeInstConfMap := make(map[string]interface{})

	for index, _ := range runningActiveMgr.RunningActives {
		runningActive := &runningActiveMgr.RunningActives[index]

		activeType := req.ActiveType
		activeID := runningActive.ActiveID

		if runningActive.ActiveType == activeType &&
			runningActive.ChannelID == req.ChannelID &&
			!runningActive.IsExpired() &&
			runningActive.ActiveType != proto.E_ACTIVETYPE_CDKEYEXCHANGE {
			base.GLog.Debug("Uin[%d] ZoneID[%d] ActiveType[%s] ActiveID[%d] ChannneID[%s] is running",
				req.Uin, req.ZoneID, activeType.String(),
				activeID, req.ChannelID)

			// 获取活动配置
			activeInstConfigKey := fmt.Sprintf(proto.ActiveInstConfigKeyFmt, req.ZoneID, activeType.String(),
				activeID)
			activeInstConf, err := getActiveInstConfig(activeType, activeInstConfigKey)
			if err != nil {
				continue
			}
			activeInstBaseConf := getActiveInstBaseConfig(activeType, activeInstConf)
			if activeInstBaseConf == nil {
				continue
			}

			// 设置配置
			activeInstConfMap[activeInstConfigKey] = activeInstConf

			activeStartTime := runningActive.BeginTime()
			activeEndTime := runningActive.EndTime()

			// 查询玩家的活动信息
			playerActiveInfo := proto.QueryPlayerActiveInfo(req.Uin, activeType, activeID, common.GDBEngine)
			if playerActiveInfo == nil {
				base.GLog.Debug("Uin[%d] first contract ActiveType[%s] ActiveID[%d]", req.Uin, activeType.String(), activeID)
				// 插入
				activeResetTime := activeEndTime
				if activeInstBaseConf.ResetEveryDay == 1 {
					// 重置时间设置为24小时
					activeResetTime = activeStartTime.Add(common.DayDuration)
				}

				// 活跃任务有自己的tasktype
				taskType := proto.DEFAULT_TASKTYPE
				if activeType == proto.E_ACTIVETYPE_MISSION {
					taskType = activeInstConf.(*proto.ActiveMissionInfoS).TaskType
				}

				// 创建记录
				playerActiveInfo = proto.CreatePlayerActiveInfo(req.Uin, req.ZoneID, activeType, activeID, req.ChannelID,
					&activeStartTime, &activeEndTime, &activeResetTime, taskType, 0, common.GDBEngine)
				if playerActiveInfo == nil {
					err = fmt.Errorf("Uin[%d] ActiveType[%s] ActiveID[%d] create record failed!",
						req.Uin, activeType.String(), activeID)
					base.GLog.Error(err.Error())
					return nil, err
				}
			} else {
				// 判断活动已经存在，但已过期
				if now.After(playerActiveInfo.ActiveEndTime) {
					// 清理玩家数据
					base.GLog.Debug("Uin[%d] ActiveType[%s] ActiveID[%d] Reset", req.Uin, activeType.String(),
						activeID)
					activeResetTime := activeEndTime
					if activeInstBaseConf.ResetEveryDay == 1 {
						activeResetTime = activeStartTime.Add(common.DayDuration)
					}
					err = playerActiveInfo.ResetActive(&activeStartTime, &activeEndTime, &activeResetTime, common.GDBEngine)
					if err != nil {
						continue
					}
				}
			}

			playerActiveInfos = append(playerActiveInfos, playerActiveInfo)
		}
	}

	res.Uin = req.Uin
	res.ZoneID = req.ZoneID
	res.ActiveType = req.ActiveType
	// 获取具体每个实例的完成情况
	err = getPlayerActiveInstancesInfo(req, playerActiveInfos, activeInstConfMap, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func getPlayerActiveInstancesInfo(req *proto.ProtoGetPlayerActiveReq, playerActiveInfos []*proto.TblPlayerActiveInfo,
	activeInstConfMap map[string]interface{}, res *proto.ProtoGetPlayerActiveRes) error {
	var err error
	now := time.Now()

	res.ActiveInstanceLst = make([]proto.ProtoActiveInstanceInfo, len(playerActiveInfos))

	for index, playerActiveInfo := range playerActiveInfos {
		activeType := req.ActiveType
		activeID := playerActiveInfo.ActiveID

		// 查询活动静态配置
		activeInstConfigKey := fmt.Sprintf(proto.ActiveInstConfigKeyFmt, req.ZoneID, activeType.String(),
			activeID)

		activeInstConf, isExist := activeInstConfMap[activeInstConfigKey]
		if !isExist {
			continue
		}

		activeInstBaseConfig := getActiveInstBaseConfig(activeType, activeInstConf)
		if activeInstBaseConfig == nil {
			continue
		}

		base.GLog.Debug("Uin[%d] ZoneID[%d] ActiveType[%s] ActiveID[%d] now[%s] ActiveStartTime[%s] ActiveResetTime[%s] ActiveEndTime[%s]",
			playerActiveInfo.Uin, playerActiveInfo.ZoneID, playerActiveInfo.ActiveType.String(), playerActiveInfo.ActiveID,
			base.TimeName(now),
			base.TimeName(playerActiveInfo.ActiveStartTime), base.TimeName(playerActiveInfo.ActiveResetTime),
			base.TimeName(playerActiveInfo.ActiveEndTime))

		// 如果时间已经超过重置时间
		if now.After(playerActiveInfo.ActiveResetTime) && now.Before(playerActiveInfo.ActiveEndTime) {
			// 重置
			playerActiveInfo.AccumulateCount = 0
			playerActiveInfo.ReceiveCount = 0
			// 计算下次的重置时间
			for true {
				playerActiveInfo.ActiveResetTime = playerActiveInfo.ActiveResetTime.Add(common.DayDuration)
				if playerActiveInfo.ActiveResetTime.After(now) {
					break
				}
			}
			base.GLog.Warn("Uin[%d] ZoneID[%d] ActiveType[%s] ActiveID[%d] Next-ActiveResetTime[%s] Time Exceed will Reset",
				req.Uin, req.ZoneID, activeType.String(), activeID, base.TimeName(playerActiveInfo.ActiveResetTime))
			playerActiveInfo.SyncDB(common.GDBEngine)
		}

		// 存放活动配置
		res.ActiveInstanceLst[index].ActiveinstanceConfig = activeInstConf
		res.ActiveInstanceLst[index].ActiveID = activeID
		// 活动剩余时间
		res.ActiveInstanceLst[index].RemainderSeconds = int64(playerActiveInfo.ActiveEndTime.Sub(now).Seconds())
		// 玩家在该活动的累积数量，超级礼包该字段无效，兑换活动是作为领取的判断条件
		res.ActiveInstanceLst[index].AccumulateCount = playerActiveInfo.AccumulateCount
		// 领取的次数，有的可以领取多次，有的一天只能领取一次
		res.ActiveInstanceLst[index].ReceiveCount = playerActiveInfo.ReceiveCount
		// 计算刷新剩余秒数
		res.ActiveInstanceLst[index].RefreshRemainderSecs = int64(playerActiveInfo.ActiveResetTime.Sub(now).Seconds())
	}
	return err
}

// 购买超值礼包后发货
func receiveSuperGiftActive(req *proto.ProtoDeliveryRechargeCommodityReq, res *proto.ProtoDeliveryRechargeCommodityRes) error {
	var err error
	base.GLog.Debug("Uin[%d] ZoneID[%d] receive SuperGift[%d]", req.Uin, req.ZoneID, req.RechargeCommdityID)

	activeID := req.RechargeCommdityID
	activeInstConfigKey := fmt.Sprintf(proto.ActiveInstConfigKeyFmt, req.ZoneID, proto.E_ACTIVETYPE_SUPERGIFTPACKAGE.String(),
		activeID)

	// 超级礼包的配置
	activeInstConfig := new(proto.ActiveSuperGiftInfoS)
	err = common.GetStrDataFromRedis(activeInstConfigKey, activeInstConfig)
	if err != nil {
		return err
	}
	base.GLog.Debug("SuperGift[%d] %+v", activeID, activeInstConfig)

	// 查询玩家的活动信息
	playerActiveInfo := proto.QueryPlayerActiveInfo(req.Uin, proto.E_ACTIVETYPE_SUPERGIFTPACKAGE, activeID, common.GDBEngine)
	if playerActiveInfo == nil {
		err = fmt.Errorf("Uin[%d] ActiveInfo is not exist! recevie supergift", req.Uin)
		base.GLog.Error(err.Error())
		return err
	}

	if playerActiveInfo.ReceiveCount < activeInstConfig.Base.ReceiveLimit {
		playerActiveInfo.ReceiveCount++
		base.GLog.Debug("Uin[%d] ZoneID[%d] ActiveType[%s] ActiveID[%d] ReceiveCount[%d]",
			req.Uin, req.ZoneID, proto.E_ACTIVETYPE_SUPERGIFTPACKAGE.String(), activeID, playerActiveInfo.ReceiveCount)
	} else {
		err = fmt.Errorf("Uin[%d] ZoneID[%d] ActiveType[%s] ActiveID[%d] ReceiveCount[%d] exceed ReceiveLimitCount[%d]",
			req.Uin, req.ZoneID, proto.E_ACTIVETYPE_SUPERGIFTPACKAGE.String(), activeID, playerActiveInfo.ReceiveCount,
			activeInstConfig.Base.ReceiveLimit)
		base.GLog.Error(err.Error())
		return err
	}

	var rechargeNtf csssvr_proto.ProtoCssSvrUserRechargeNtf
	rechargeNtf.ChannelID = req.ChannelID
	rechargeNtf.PlatForm = req.PlatForm
	rechargeNtf.RechargeAmount = activeInstConfig.Price
	rechargeNtf.Uin = req.Uin
	consul_api.PostRequstByConsulDns(req.Uin, "CssSvrUserRecharge", &rechargeNtf, common.ConsulClient, "CassandraSvr")

	// 同步到玩家活动数据库
	playerActiveInfo.SyncDB(common.GDBEngine)
	res.InnerGoods = activeInstConfig.InnerGoods
	res.NameKey = activeInstConfig.NameKey
	return nil
}

func getRunningActives(zoneID int32) (*proto.RunningActiveMgr, error) {
	runningActiveMgr := getActiveControlConfig(zoneID)
	if runningActiveMgr == nil {
		// key不存在，需要初始化下
		runningActiveMgr = proto.CreateRunningActiveMgr(zoneID, common.GRedis)
	}

	if runningActiveMgr.IsEmpty() {
		err := fmt.Errorf("ZoneID[%d] There is no open activties!", zoneID)
		base.GLog.Error(err.Error())
		return nil, err
	}
	//base.GLog.Debug("Now open active:+%v", runningActiveMgr.RunningActives)
	return runningActiveMgr, nil
}

// 活动运行时玩家数据累计
func ActiveAccumulateParameterNtf(req *proto.ProtoActiveAccumulateParameterNtf) (*proto.ProtoActiveCanReceiveNtf, error) {
	var err error
	var res proto.ProtoActiveCanReceiveNtf
	res.Uin = req.Uin
	res.ZoneID = req.ZoneID
	res.ActiveType = req.ActiveType

	now := time.Now()
	activeType := req.ActiveType

	// 查询开放的活动
	runningActiveMgr, err := getRunningActives(req.ZoneID)
	if err != nil {
		return nil, err
	}

	// 开放活跃任务活动id集合
	runningActiveMissionIDSet := hashset.New()
	// 开放活跃任务的控制信息
	runningActiveMissionCtrlInfos := make([]*proto.ProtoActiveControlInfoS, 0)
	// 开放活跃任务的配置信息
	runningActiveMissionConfigInfos := make([]*proto.ActiveMissionInfoS, 0)

	var runningActiveCtrlInfo *proto.ProtoActiveControlInfoS
	for index := range runningActiveMgr.RunningActives {
		runningActiveCtrlInfo = &runningActiveMgr.RunningActives[index]

		if runningActiveCtrlInfo.ActiveType == proto.E_ACTIVETYPE_MISSION &&
			!runningActiveCtrlInfo.IsExpired() {
			runningActiveMissionCtrlInfos = append(runningActiveMissionCtrlInfos, runningActiveCtrlInfo)

			base.GLog.Debug("runningActiveCtrlInfo:%+v", runningActiveCtrlInfo)

			activeInstConfigKey := fmt.Sprintf(proto.ActiveInstConfigKeyFmt, req.ZoneID, proto.E_ACTIVETYPE_MISSION.String(),
				runningActiveCtrlInfo.ActiveID)

			activeInstConf, err := getActiveInstConfig(proto.E_ACTIVETYPE_MISSION, activeInstConfigKey)
			if err == nil {
				runningActiveMissionConfigInfo := activeInstConf.(*proto.ActiveMissionInfoS)
				runningActiveMissionConfigInfos = append(runningActiveMissionConfigInfos, runningActiveMissionConfigInfo)
				base.GLog.Debug("runningActiveMissionConfigInfo:%+v", runningActiveMissionConfigInfo)
			}
		}
	}

	//-------------------------------------

	for index := range req.ActiveDatas {
		activeNtfData := &req.ActiveDatas[index]
		taskType := activeNtfData.ActiveTaskType
		taskOpType := activeNtfData.TaskOpType
		AccumalateParamter := activeNtfData.AccumalateParamter

		base.GLog.Debug("Uin[%d] ZoneID[%d] ActiveType[%s] activeNtfData:%+v",
			req.Uin, req.ZoneID, proto.E_ACTIVETYPE_MISSION.String(), activeNtfData)

		// 清空
		runningActiveMissionIDSet.Clear()
		// 判断该活动是否在在开放列表中，这里只有通过tasktype来对比
		activeMissionIsRunning := false
		for index := range runningActiveMissionConfigInfos {
			if taskType == runningActiveMissionConfigInfos[index].TaskType {
				activeMissionIsRunning = true
				// 开放活动中，tasktype相同的活跃任务id
				runningActiveMissionIDSet.Add(runningActiveMissionConfigInfos[index].Base.ActiveID)
			}
		}

		if !activeMissionIsRunning {
			base.GLog.Debug("Uin[%d] ZoneID[%d] ActiveType[%s] activeNtfData:%+v is not open!",
				req.Uin, req.ZoneID, proto.E_ACTIVETYPE_MISSION.String(), activeNtfData)
			continue
		}

		idsJson, _ := runningActiveMissionIDSet.ToJSON()
		base.GLog.Debug("%s:%s open ids[%s]", proto.E_ACTIVETYPE_MISSION.String(), taskType, string(idsJson))

		// 查询玩家参与的活跃任务记录，根据tasktype条件查询
		playerMissionActiveInfos, err := proto.QueryPlayerMissionActionInfos(req.Uin, taskType, common.GDBEngine)
		if err == nil {
			base.GLog.Debug("Uin[%d] ZoneID[%d] ActiveType[%s] taskType[%s] recordCount[%d]",
				req.Uin, req.ZoneID, proto.E_ACTIVETYPE_MISSION.String(), taskType, len(playerMissionActiveInfos))

			for index := range playerMissionActiveInfos {
				missionActive := &playerMissionActiveInfos[index]
				if activeOpen := runningActiveMissionIDSet.Contains(missionActive.ActiveID); activeOpen == true {
					// 该活动是开放的
					base.GLog.Debug("Uin[%d] ActiveType[%s] ActiveID[%d] taskType[%s] is open",
						req.Uin, activeType.String(), missionActive.ActiveID, taskType)

					// 活动已经过期
					if now.After(missionActive.ActiveEndTime) {
						err = fmt.Errorf("Uin[%d] ZoneID[%d] ActiveType[%s] ActiveID[%d] has expired! ActiveEndTime[%s] now[%s]",
							req.Uin, req.ZoneID, activeType.String(), missionActive.ActiveID, base.TimeName(missionActive.ActiveEndTime),
							base.TimeName(now))
						base.GLog.Error(err.Error())
						continue
					}

					// 如果找到了，就从列表中删除掉
					runningActiveMissionIDSet.Remove(missionActive.ActiveID)

					// 修改玩家数据
					if taskOpType == 1 {
						missionActive.AccumulateCount = AccumalateParamter
						missionActive.SyncDB(common.GDBEngine)
					} else if taskOpType == 2 {
						missionActive.AccumulateCount += AccumalateParamter
						missionActive.SyncDB(common.GDBEngine)
					} else {
						base.GLog.Error("Uin[%d] ActiveType[%s] ActiveID[%d] taskType[%s] taskOpType[%d] is invalid!",
							req.Uin, activeType.String(), missionActive.ActiveID, taskType, taskOpType)
						continue
					}

					// 判断活跃任务是否满足领取条件
					activeInstConfigKey := fmt.Sprintf(proto.ActiveInstConfigKeyFmt, req.ZoneID, activeType.String(),
						missionActive.ActiveID)
					base.GLog.Debug("Uin[%d] ZoneID[%d] activeInstConfigKey[%s]", req.Uin, req.ZoneID, activeInstConfigKey)

					activeInstConf, err := getActiveInstConfig(activeType, activeInstConfigKey)
					if err != nil {
						continue
					}
					activeInstBaseConfig := getActiveInstBaseConfig(activeType, activeInstConf)
					if activeInstBaseConfig == nil {
						continue
					}

					if missionActive.AccumulateCount >= activeInstBaseConfig.ReceiveCond &&
						missionActive.ReceiveCount < activeInstBaseConfig.ReceiveLimit {
						base.GLog.Debug("Uin[%d] ActiveType[%s] ActiveID[%d] taskType[%s] can receive reward!",
							req.Uin, activeType.String(), missionActive.ActiveID, taskType)
						res.IsDone = 1
					}

				} else {
					base.GLog.Error("Uin[%d] ActiveType[%s] ActiveID[%d] taskType[%s] taskOpType[%d] is not open!",
						req.Uin, activeType.String(), missionActive.ActiveID, taskType, taskOpType)
				}
			}
		}

		// 如果活动id还有剩余的，需要创建记录
		if !runningActiveMissionIDSet.Empty() {
			iActiveMissionIDs := runningActiveMissionIDSet.Values()
			for index := range iActiveMissionIDs {
				activeMissionID := iActiveMissionIDs[index].(int)

				var amCtrl *proto.ProtoActiveControlInfoS
				var amCfg *proto.ActiveMissionInfoS
				// 找到控制信息
				for cindex := range runningActiveMissionCtrlInfos {
					if activeMissionID == runningActiveMissionCtrlInfos[cindex].ActiveID {
						amCtrl = runningActiveMissionCtrlInfos[cindex]
					}
				}
				// 找到配置信息
				for iindex := range runningActiveMissionConfigInfos {
					if activeMissionID == runningActiveMissionConfigInfos[iindex].Base.ActiveID {
						amCfg = runningActiveMissionConfigInfos[iindex]
					}
				}

				if amCfg != nil && amCtrl != nil {
					base.GLog.Debug("Uin[%d] first contract ActiveType[%s] ActiveID[%d]", req.Uin, proto.E_ACTIVETYPE_MISSION.String(), activeMissionID)

					activeStartTime := amCtrl.BeginTime()
					activeEndTime := amCtrl.EndTime()
					activeResetTime := activeEndTime

					if amCfg.Base.ResetEveryDay == 1 {
						// 重置时间设置为24小时
						activeResetTime = activeStartTime.Add(common.DayDuration)
					}

					// 计算reset时间
					if now.After(activeResetTime) && now.Before(activeEndTime) {
						// 计算下次的重置时间
						for true {
							activeResetTime = activeResetTime.Add(common.DayDuration)
							if activeResetTime.After(now) {
								break
							}
						}
					}

					// 创建记录
					playerActiveInfo := proto.CreatePlayerActiveInfo(req.Uin, req.ZoneID, proto.E_ACTIVETYPE_MISSION,
						activeMissionID, "NOAREA",
						&activeStartTime, &activeEndTime, &activeResetTime, taskType, AccumalateParamter, common.GDBEngine)
					if playerActiveInfo == nil {
						err = fmt.Errorf("Uin[%d] ActiveType[%s] ActiveID[%d] create record failed!",
							req.Uin, proto.E_ACTIVETYPE_MISSION.String(), activeMissionID)
						base.GLog.Error(err.Error())
					}
				} else {
					base.GLog.Error("Uin[%d] ActiveType[%s] ActiveID[%d] can't find control or config info",
						req.Uin, proto.E_ACTIVETYPE_MISSION.String(), activeMissionID)
				}
			}
		}
	}
	return &res, err
}

// 玩家领取活动
func PlayerActiveReceive(req *proto.ProtoPlayerActiveReceiveReq) (*proto.ProtoPlayerActiveReceiveRes, error) {
	var err error

	activeType := req.ActiveType
	activeID := req.ActiveID

	// 查询玩家活动信息
	playerActiveInfo := proto.QueryPlayerActiveInfo(req.Uin, activeType, activeID, common.GDBEngine)
	if playerActiveInfo == nil {
		err = fmt.Errorf("Uin[%d] ZoneID[%d] ActiveType[%s] ActiveID[%d] ActiveInfo is not exist!", req.Uin,
			req.ZoneID, activeType.String(), activeID)
		base.GLog.Error(err.Error())
		return nil, err
	}

	//TODO: 判断活动是否已经过期，过期就直接返回
	now := time.Now()
	if now.After(playerActiveInfo.ActiveEndTime) {
		err = fmt.Errorf("Uin[%d] ZoneID[%d] ActiveType[%s] ActiveID[%d] has expired! ActiveEndTime[%s] now[%s]",
			req.Uin, req.ZoneID, activeType.String(), activeID, base.TimeName(playerActiveInfo.ActiveEndTime),
			base.TimeName(now))
		base.GLog.Error(err.Error())
		return nil, err
	}

	activeInstConfigKey := fmt.Sprintf(proto.ActiveInstConfigKeyFmt, req.ZoneID, activeType.String(),
		activeID)
	base.GLog.Debug("Uin[%d] ZoneID[%d] activeInstConfigKey[%s]", req.Uin, req.ZoneID, activeInstConfigKey)

	// 领取的奖励
	var innerGoods string

	activeInstConf, err := getActiveInstConfig(activeType, activeInstConfigKey)
	if err != nil {
		return nil, err
	}
	activeInstBaseConfig := getActiveInstBaseConfig(activeType, activeInstConf)
	if activeInstBaseConfig == nil {
		return nil, fmt.Errorf("ActiveType[%d] is invalid!", activeType)
	}

	switch activeType {
	case proto.E_ACTIVETYPE_MISSION:
		innerGoods = activeInstConf.(*proto.ActiveMissionInfoS).InnerGoods
	case proto.E_ACTIVETYPE_EXCHANGE:
		innerGoods = activeInstConf.(*proto.ActiveExchangeInfoS).InnerGoods
	}

	// 判断是否满足领取条件
	if playerActiveInfo.AccumulateCount < activeInstBaseConfig.ReceiveCond {
		err = fmt.Errorf("Uin[%d] ZoneID[%d] ActiveType[%s] ActiveID[%d] AccumulateCount[%d] less than ReceiveCond[%d]",
			req.Uin, req.ZoneID, activeType.String(),
			activeID, playerActiveInfo.AccumulateCount, activeInstBaseConfig.ReceiveCond)
		base.GLog.Error(err.Error())
		return nil, err
	}

	// 判断是否达到领取次数上限
	if playerActiveInfo.ReceiveCount >= activeInstBaseConfig.ReceiveLimit {
		err = fmt.Errorf("Uin[%d] ZoneID[%d] ActiveType[%s] ActiveID[%d] ReceiveCount[%d] reached the ReceiveLimit[%d]",
			req.Uin, req.ZoneID, activeType.String(),
			activeID, playerActiveInfo.ReceiveCount, activeInstBaseConfig.ReceiveLimit)
		base.GLog.Error(err.Error())
		return nil, err
	}

	// 领取后，累积需要清空
	playerActiveInfo.AccumulateCount = 0
	// 领取的计数递增
	playerActiveInfo.ReceiveCount++
	base.GLog.Debug("Uin[%d] ZoneID[%d] ActiveType[%s] activeID[%d] AccumulateCount[%d] ReceiveCount[%d]", req.Uin, req.ZoneID,
		req.ActiveType.String(), req.ActiveID, playerActiveInfo.AccumulateCount, playerActiveInfo.ReceiveCount)
	playerActiveInfo.SyncDB(common.GDBEngine)

	var res proto.ProtoPlayerActiveReceiveRes
	res.Uin = req.Uin
	res.ZoneID = req.ZoneID
	res.ActiveType = activeType
	res.ActiveID = activeID
	res.AccumulateCount = playerActiveInfo.AccumulateCount
	res.ReceiveCount = playerActiveInfo.ReceiveCount
	res.InnerGoods = innerGoods

	return &res, nil
}

func GetActiveExchangeCost(req *proto.ProtoGetActiveExchangeCostReq) (*proto.ProtoGetActiveExchangeCostRes, error) {
	var err error
	var res proto.ProtoGetActiveExchangeCostRes

	activeType := req.ActiveType
	activeID := req.ActiveID

	activeInstConfigKey := fmt.Sprintf(proto.ActiveInstConfigKeyFmt, req.ZoneID, activeType.String(),
		activeID)
	base.GLog.Debug("Uin[%d] ZoneID[%d] activeInstConfigKey[%s]", req.Uin, req.ZoneID, activeInstConfigKey)

	activeInstConfig := new(proto.ActiveExchangeInfoS)
	err = common.GetStrDataFromRedis(activeInstConfigKey, activeInstConfig)
	if err != nil {
		return nil, err
	}
	base.GLog.Debug("ActiveMission[%d] %+v", activeID, activeInstConfig)

	res.Uin = req.Uin
	res.ZoneID = req.ZoneID
	res.ActiveID = req.ActiveID
	res.ActiveType = res.ActiveType
	res.ExchangeCost = activeInstConfig.ExchangeCost
	return &res, nil
}

func PlayerExchangeCDKey(req *proto.ProtoPlayerExchangeCDKeyReq) (*proto.ProtoPlayerExchangeCDKeyRes, error) {
	var err error
	var res proto.ProtoPlayerExchangeCDKeyRes
	res.Uin = req.Uin
	res.ZoneID = req.ZoneID

	// 开放的活动列表
	runningActiveMgr := getActiveControlConfig(req.ZoneID)
	if runningActiveMgr == nil {
		// key不存在，需要初始化下
		runningActiveMgr = proto.CreateRunningActiveMgr(req.ZoneID, common.GRedis)
	}

	if runningActiveMgr.IsEmpty() {
		err = fmt.Errorf("Uin[%d] ZoneID[%d] is currently no open active!", req.Uin, req.ZoneID)
		base.GLog.Error(err)
		return nil, err
	}

	// 过滤出所有cdkey兑换活动
	runningCDKeyExchangeActiveIDLst := singlylinkedlist.New()
	var runningActive *proto.ProtoActiveControlInfoS
	for index := range runningActiveMgr.RunningActives {
		runningActive = &runningActiveMgr.RunningActives[index]
		if runningActive.ActiveType == proto.E_ACTIVETYPE_CDKEYEXCHANGE {
			runningCDKeyExchangeActiveIDLst.Add(runningActive.ActiveID)
		}
	}

	// 查询玩家cdkey兑换信息
	playerCDKeyExchangeRecord := proto.QueryPlayerCDKeyExchange(req.Uin, common.GDBEngine)
	if playerCDKeyExchangeRecord == nil {
		// 玩家第一次使用cdkey兑换
		base.GLog.Debug("Uin[%d] first contract ActiveCDKeyExchange", req.Uin)
		playerCDKeyExchangeRecord = proto.CreatePlayerCDKeyExchange(req.Uin, req.ZoneID, common.GDBEngine)
		if playerCDKeyExchangeRecord == nil {
			err = fmt.Errorf("Uin[%d] ActiveType[%s] create record failed!", req.Uin, proto.E_ACTIVETYPE_CDKEYEXCHANGE.String())
			base.GLog.Error(err.Error())
			return nil, err
		}
	}

	idsJson, _ := runningCDKeyExchangeActiveIDLst.ToJSON()
	base.GLog.Debug("%s open ids[%s]", proto.E_ACTIVETYPE_CDKEYEXCHANGE, string(idsJson))

	// 通过id 拿到静态配置
	exchangeOk := runningCDKeyExchangeActiveIDLst.Any(func(index int, value interface{}) bool {
		activeID := value.(int)
		activeInstConfigKey := fmt.Sprintf(proto.ActiveInstConfigKeyFmt, req.ZoneID, proto.E_ACTIVETYPE_CDKEYEXCHANGE.String(),
			activeID)
		//base.GLog.Debug("Uin[%d] ZoneID[%d] activeInstConfigKey[%s]", req.Uin, req.ZoneID, activeInstConfigKey)
		activeInstConf, err := getActiveInstConfig(proto.E_ACTIVETYPE_CDKEYEXCHANGE, activeInstConfigKey)
		if err == nil {
			cdkeyExchangeConf := activeInstConf.(*proto.ActiveCDKeyExchangeInfoS)
			if cdkeyExchangeConf.CDKey == req.CDKey {
				// 配置的cdkey是运行，且和玩家输入一致，判断玩家是否领取过
				exchangeRes := playerCDKeyExchangeRecord.ExchangeCDKey(req.CDKey)
				if exchangeRes {
					base.GLog.Debug("Uin[%d] ZoneID[%d] activeInstConfigKey[%s] CDKey[%s] exchange successed!", req.Uin, req.ZoneID,
						activeInstConfigKey, req.CDKey)
					playerCDKeyExchangeRecord.SyncDB(common.GDBEngine)
					res.InnerGoods = cdkeyExchangeConf.InnerGoods

					// cdkey领取计数递增
					cdkeyCountKey := fmt.Sprintf(proto.CDKeyCountKeyFmt, activeInstConfigKey)
					common.GRedis.Incr(cdkeyCountKey)
				} else {
					base.GLog.Debug("Uin[%d] ZoneID[%d] activeInstConfigKey[%s] CDKey[%s] already exchanged!", req.Uin, req.ZoneID,
						activeInstConfigKey, req.CDKey)
					return false
				}
				return true
			}
		}
		return false
	})

	if !exchangeOk {
		err = fmt.Errorf("Uin[%d] ZoneID[%d] %s CDKey[%s] exchange failed!", req.Uin, req.ZoneID, proto.E_ACTIVETYPE_CDKEYEXCHANGE.String(),
			req.CDKey)
		base.GLog.Error(err.Error())
		return nil, err
	}

	return &res, nil
}

func CheckPlayerActiveIsCompleted(req *proto.ProtoCheckPlayerActiveIsCompletedReq) (*proto.ProtoCheckPlayerActiveIsCompletedRes, error) {
	var err error
	var res proto.ProtoCheckPlayerActiveIsCompletedRes
	res.Uin = req.Uin
	res.ZoneID = req.ZoneID

	userFinance, _ := QueryFinanceUser(req.Uin)
	if userFinance == nil {
		err = fmt.Errorf("Uin[%d] is not exist!", req.Uin)
		base.GLog.Error(err.Error())
		return nil, err
	}

	var isCompleted int32
	if userFinance.FirstRecharge.ReceiveCount == 0 {
		isCompleted = 1
	}

	res.ActiveCompleteLst = append(res.ActiveCompleteLst, proto.ActiveIsCompleteS{
		ActiveType:  proto.E_ACTIVETYPE_FIRSTRECHARGE,
		IsCompleted: isCompleted,
	})

	newPlayerLoginBenefits, _ := QueryNewPlayerLoginBenefit(req.Uin)
	if newPlayerLoginBenefits == nil {
		err = fmt.Errorf("Uin[%d] is not exist!", req.Uin)
		base.GLog.Error(err.Error())
		return nil, err
	}

	res.ActiveCompleteLst = append(res.ActiveCompleteLst, proto.ActiveIsCompleteS{
		ActiveType:  proto.E_ACTIVETYPE_NEWPLAYERBENEFIT,
		IsCompleted: newPlayerLoginBenefits.IsCompleted,
	})

	return &res, nil
}
