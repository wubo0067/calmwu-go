/*
 * @Author: calmwu
 * @Date: 2018-07-17 10:36:26
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-07-17 10:40:58
 * @Comment:
 */

package web

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	"sailcraft/base"
	"sailcraft/csssvr_main/common"
	"sailcraft/csssvr_main/proto"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

const (
	//lilith公钥
	LILITH_PUBLIC_KEY = `
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC6l0l8aJHIPB5thuz8gJxof0i+
d6tJ/MXO/BHVy4iSdf9zbaPu29vyYmJXRHqEEZ4LexvJr/miTVj/fj+nEY9qHTtK
Y0f5a6K0iKjF6AF93tPDyu+IOSqyJ/j7OQLtaeRftC5FiNBY1vqI0PXzPaDbztFW
UKRMxOXCLt4pVwp/UQIDAQAB
-----END PUBLIC KEY-----
`
)

func (cs *CassandraSvrModule) UploadBattleVideo(c *gin.Context) {
	webItfProcess(c, "UploadBattleVideo", INTERFACE_OPTYPE_PUT)
}

func (cs *CassandraSvrModule) DeleteBattleVideo(c *gin.Context) {
	webItfProcess(c, "DeleteBattleVideo", INTERFACE_OPTYPE_PUT)
}

func (cs *CassandraSvrModule) GetBattleVideo(c *gin.Context) {
	webItfProcess(c, "GetBattleVideo", INTERFACE_OPTYPE_GET)
}

func (cs *CassandraSvrModule) OldUserReceiveCompensation(c *gin.Context) {
	webItfProcess(c, "OldUserReceiveCompensation", INTERFACE_OPTYPE_GET)
}

func (cs *CassandraSvrModule) TuitionStepReport(c *gin.Context) {
	webItfProcess(c, "TuitionStepReport", INTERFACE_OPTYPE_PUT)
}

func (cs *CassandraSvrModule) CssSvrUserLogin(c *gin.Context) {
	webItfProcess(c, "CssSvrUserLogin", INTERFACE_OPTYPE_PUT)
}

func (cs *CassandraSvrModule) CssSvrUserLogout(c *gin.Context) {
	webItfProcess(c, "CssSvrUserLogout", INTERFACE_OPTYPE_PUT)
}

func (cs *CassandraSvrModule) CssSvrUserRecharge(c *gin.Context) {
	webItfProcess(c, "CssSvrUserRecharge", INTERFACE_OPTYPE_PUT)
}

func (cs *CassandraSvrModule) SvrQueryPlayerGeoInfo(c *gin.Context) {
	webItfProcess(c, "SvrQueryPlayerGeoInfo", INTERFACE_OPTYPE_GET)
}

func (cs *CassandraSvrModule) UploadUserAction(c *gin.Context) {
	webItfProcess(c, "UploadUserAction", INTERFACE_OPTYPE_PUT)
}

func (cs *CassandraSvrModule) QueryUserRechargeInfo(c *gin.Context) {
	webItfProcess(c, "QueryUserRechargeInfo", INTERFACE_OPTYPE_GET)
}

func (cs *CassandraSvrModule) ClientCDNResourceDownloadReport(c *gin.Context) {
	webItfProcess(c, "ClientCDNResourceDownloadReport", INTERFACE_OPTYPE_PUT)
}

func (cs *CassandraSvrModule) QueryCountryISOCode(c *gin.Context) {
	req := base.UnpackRequest(c)
	if req == nil {
		base.GLog.Error("SvrQueryPlayerGeoInfo read request failed!")
		return
	}

	var reqData proto.ProtoQueryCountryISOByIpReq
	err := mapstructure.Decode(req.ReqData.Params, &reqData)
	if err != nil {
		base.GLog.Error("Uin[%d] Decode ProtoQueryCountryISOByIpReq failed! reason[%s]",
			req.Uin, err.Error())
		return
	}

	var res base.ProtoResponseS
	res.Version = req.Version
	res.TimeStamp = time.Now().Unix()
	res.EventId = req.EventId
	res.ReturnCode = 0
	res.ResData.InterfaceName = req.ReqData.InterfaceName

	var resData proto.ProtoQueryCountryISOByIpRes
	resData.Uin = reqData.Uin
	resData.ClientIP = reqData.ClientIP
	resData.CountryISOCode, resData.CountryName = common.QueryGeoInfo(reqData.ClientIP)

	res.ResData.Params = resData
	base.GLog.Debug("QueryCountryISOCode response to uin[%d]", req.Uin)
	base.SendResponse(c, &res)
}

func (cs *CassandraSvrModule) LilithVerifySign(c *gin.Context) {
	req := base.UnpackRequest(c)
	if req == nil {
		base.GLog.Error("LilithVerifySign read request failed!")
		return
	}

	var reqData proto.ProtoVerifySignReq
	err := mapstructure.Decode(req.ReqData.Params, &reqData)
	if err != nil {
		base.GLog.Error("Uin[%d] Decode ProtoVerifySignReq failed! reason[%s]",
			req.Uin, err.Error())
		return
	}

	block, _ := pem.Decode([]byte(LILITH_PUBLIC_KEY))
	if block == nil {
		base.GLog.Error("pem.Decode LILITH_PUBLIC_KEY failed!")
		return
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		base.GLog.Error("Failed to parse RSA public key: %s", err)
		return
	}
	rsaPub, _ := pub.(*rsa.PublicKey)

	var res base.ProtoResponseS
	res.Version = req.Version
	res.TimeStamp = time.Now().Unix()
	res.EventId = req.EventId
	res.ReturnCode = 0
	res.ResData.InterfaceName = req.ReqData.InterfaceName

	var resData proto.ProtoVerifySignRes
	resData.Serial = reqData.Serial
	resData.VerifyResult = -1

	h := sha1.New()
	h.Write([]byte(reqData.SignString))
	digest := h.Sum(nil)

	//ds, _ := base64.StdEncoding.DecodeString(reqData.AuthString)

	err = rsa.VerifyPKCS1v15(rsaPub, crypto.SHA1, digest, []byte(reqData.AuthString)) //ds)
	if err != nil {
		base.GLog.Error("Serial[%s] VerifyPKCS1v15 failed! reason[%s]",
			reqData.Serial, err.Error())
	} else {
		resData.VerifyResult = 0
	}

	res.ResData.Params = resData

	base.SendResponse(c, &res)
}
