package handler

import (
	"sailcraft/fleetsvr_main/model"
	"sailcraft/fleetsvr_main/table"
)

func GetMultiUserHeadFrame(uinSlice ...int) (map[int]*table.TblUserHeadFrame, error) {
	return model.GetMultiUserHeadFrame(uinSlice...)
}

func GetUserHeadFrame(uin int) (*table.TblUserHeadFrame, error) {
	return model.GetUserHeadFrame(uin)
}
