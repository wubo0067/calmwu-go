/*
 * @Author: calmwu
 * @Date: 2018-11-19 16:47:34
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-19 16:52:08
 */

package main

import (
	base "doyo-server-go/doyo-base-go"
	recdata_proto "doyo-server-go/doyo-recdatasvr-go/proto"
	routerstub "doyo-server-go/doyo-routersvr-go/doyo-routerstub"
	routersvr_proto "doyo-server-go/doyo-routersvr-go/proto"

	"github.com/pquerna/ffjson/ffjson"
)

func notifyUserLogin(doyoRsm *routerstub.RouterStubModule, svrInstanceTopic string, userID string,
	userLanguage string, userCountry string) {
	userLogin := recdata_proto.ProtoDoyoRecDataUserLogin{
		UserID:       userID,
		UserLanguage: userLanguage,
		UserCountry:  userCountry,
		UserGender:   1,
	}

	jsonLogin, _ := ffjson.Marshal(userLogin)
	appMsg := routersvr_proto.AppServMsg{
		AppServCmdID:   int(recdata_proto.DoyoRecDataCmdUserLogin),
		AppServCmdData: jsonLogin,
	}

	jsonMsg, _ := ffjson.Marshal(appMsg)
	msgID, err := routerstub.Notify(doyoRsm, svrInstanceTopic, "DoyoRecDataSvr", routersvr_proto.RouterSvrDispatchPolicyRR, "", jsonMsg)
	if err != nil {
		base.ZLog.Errorf("notify failed! reason:%s", err.Error())
		return
	}
	base.ZLog.Debugf("notifty msgid[%s]", msgID)
}

func notifyUserLogout(doyoRsm *routerstub.RouterStubModule, svrInstanceTopic string, userID string) {
	userLogout := recdata_proto.ProtoDoyoRecDataUserLogout{
		UserID: userID,
	}

	jsonLogout, _ := ffjson.Marshal(userLogout)
	appMsg := routersvr_proto.AppServMsg{
		AppServCmdID:   int(recdata_proto.DoyoRecDataCmdUserLogout),
		AppServCmdData: jsonLogout,
	}

	jsonMsg, _ := ffjson.Marshal(appMsg)
	msgID, err := routerstub.Notify(doyoRsm, svrInstanceTopic, "DoyoRecDataSvr", routersvr_proto.RouterSvrDispatchPolicyRR, "", jsonMsg)
	if err != nil {
		base.ZLog.Errorf("notify failed! reason:%s", err.Error())
		return
	}
	base.ZLog.Debugf("notifty msgid[%s]", msgID)
}

func notifyAnchorStartRoom(doyoRsm *routerstub.RouterStubModule, svrInstanceTopic string, ntf *recdata_proto.ProtoDoyoRecDataAnchorStartRoom) {
	jsonRoomStart, _ := ffjson.Marshal(ntf)
	appMsg := routersvr_proto.AppServMsg{
		AppServCmdID:   int(recdata_proto.DoyoRecDataCmdAnchorStartRoom),
		AppServCmdData: jsonRoomStart,
	}
	jsonMsg, _ := ffjson.Marshal(appMsg)
	msgID, err := routerstub.Notify(doyoRsm, svrInstanceTopic, "DoyoRecDataSvr", routersvr_proto.RouterSvrDispatchPolicyRR, "", jsonMsg)
	if err != nil {
		base.ZLog.Errorf("notify failed! reason:%s", err.Error())
		return
	}
	base.ZLog.Debugf("notifty msgid[%s]", msgID)
}
