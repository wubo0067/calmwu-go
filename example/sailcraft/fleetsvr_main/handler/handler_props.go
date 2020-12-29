package handler

import (
	"sailcraft/fleetsvr_main/config"
	"sailcraft/fleetsvr_main/model"
	"sailcraft/fleetsvr_main/proto"
	"sailcraft/fleetsvr_main/table"
)

func GetMultiPropsByProtypeId(uin int, protypeIds ...int) ([]*table.TblPropInfo, error) {
	propInfoModel := model.PropInfoModel{Uin: uin}
	return propInfoModel.GetMultiPropsByProtypeId(protypeIds...)
}

func GetPropByProtypeId(uin int, protypeId int) (*table.TblPropInfo, error) {
	protoInfoModel := model.PropInfoModel{Uin: uin}
	return protoInfoModel.GetPropByProtypeId(protypeId)
}

func composeProtoPropInfo(target *proto.ProtoPropItem, data *table.TblPropInfo) {
	target.ProtypeId = data.ProtypeId
	target.CountType = config.COUNT_TYPE_CONST
	target.Count = data.PropNum
}
