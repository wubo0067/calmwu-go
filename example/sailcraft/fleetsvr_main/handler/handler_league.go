package handler

import (
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/model"
	"sailcraft/fleetsvr_main/table"
)

func GetLeagueInfo(uin int) (*table.TblLeagueInfo, error) {
	if uin <= 0 {
		return nil, custom_errors.New("uin is invalid")
	}

	leagueModel := model.LeagueModel{Uin: uin}
	leagueInfo, err := leagueModel.GetLeagueInfo()
	if err != nil {
		return nil, err
	}

	if leagueInfo == nil {
		return nil, custom_errors.New("user league info not exist")
	}

	return leagueInfo, nil
}

func GetMultiLeagueInfo(uinSlice ...int) ([]*table.TblLeagueInfo, error) {
	return model.GetMultiLeagueInfo(uinSlice...)
}

func GetLeagueSeasonInfo() (*table.TblLeagueSeasonInfo, error) {
	leagueSeasonModel := model.LeagueSeasonModel{}
	return leagueSeasonModel.GetLeagueSeasonInfo()
}
