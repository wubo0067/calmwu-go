package handler

import (
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/table"
)

type UserInfoForClient struct {
	Uin                int
	UserInfo           *table.TblUserInfo
	LeagueInfo         *table.TblLeagueInfo
	HeadFrameProtypeId int
	HeadId             string
	HeadType           int
}

func GetUserInfoForClientList(uinList ...int) (map[int]*UserInfoForClient, error) {
	if len(uinList) <= 0 {
		return nil, custom_errors.New("uin list is empty")
	}

	infoMap := make(map[int]*UserInfoForClient)

	allUsers, err := GetMultiUserInfo(uinList...)
	if err != nil {
		return nil, err
	}

	for _, user := range allUsers {
		if userInfoForClient, ok := infoMap[user.Uin]; ok {
			userInfoForClient.UserInfo = user
		} else {
			userInfoForClient = new(UserInfoForClient)
			userInfoForClient.UserInfo = user
			userInfoForClient.Uin = user.Uin
			infoMap[user.Uin] = userInfoForClient
		}
	}

	allLeagues, err := GetMultiLeagueInfo(uinList...)
	if err != nil {
		return nil, err
	}

	for _, leagueInfo := range allLeagues {
		if userInfoForClient, ok := infoMap[leagueInfo.Uin]; ok {
			userInfoForClient.LeagueInfo = leagueInfo
		} else {
			userInfoForClient = new(UserInfoForClient)
			userInfoForClient.LeagueInfo = leagueInfo
			userInfoForClient.Uin = leagueInfo.Uin
			infoMap[leagueInfo.Uin] = userInfoForClient
		}
	}

	userHeadFrameMap, err := GetMultiUserHeadFrame(uinList...)
	if err != nil {
		return nil, err
	}

	for uin, tblUserHead := range userHeadFrameMap {
		if userInfoForClient, ok := infoMap[uin]; ok {
			userInfoForClient.HeadFrameProtypeId = tblUserHead.CurHeadFrame
			userInfoForClient.HeadId = tblUserHead.HeadId
			userInfoForClient.HeadType = tblUserHead.HeadType
		} else {
			userInfoForClient = new(UserInfoForClient)
			userInfoForClient.Uin = uin
			userInfoForClient.HeadFrameProtypeId = tblUserHead.CurHeadFrame
			userInfoForClient.HeadId = tblUserHead.HeadId
			userInfoForClient.HeadType = tblUserHead.HeadType
			infoMap[uin] = userInfoForClient
		}
	}

	return infoMap, nil
}
