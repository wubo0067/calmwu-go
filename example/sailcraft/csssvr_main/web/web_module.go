/*
 * @Author: calmwu
 * @Date: 2018-01-10 16:26:29
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-06-13 15:36:46
 * @Comment:
 */

package web

import "sailcraft/base"

type CassandraSvrModule struct {
	ModuleMetas base.WebModuleMetas
}

var (
	WebCassandraSvrModule *CassandraSvrModule = new(CassandraSvrModule)
)

func init() {
	WebCassandraSvrModule.ModuleMetas = make(base.WebModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/CassandraSvr/TuitionStepReport",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebCassandraSvrModule.TuitionStepReport, WebCassandraSvrModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/CassandraSvr/UploadBattleVideo",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebCassandraSvrModule.UploadBattleVideo, WebCassandraSvrModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/CassandraSvr/DeleteBattleVideo",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebCassandraSvrModule.DeleteBattleVideo, WebCassandraSvrModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/CassandraSvr/GetBattleVideo",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebCassandraSvrModule.GetBattleVideo, WebCassandraSvrModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/CassandraSvr/CssSvrUserLogin",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebCassandraSvrModule.CssSvrUserLogin, WebCassandraSvrModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/CassandraSvr/CssSvrUserLogout",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebCassandraSvrModule.CssSvrUserLogout, WebCassandraSvrModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/CassandraSvr/SvrQueryPlayerGeoInfo",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebCassandraSvrModule.SvrQueryPlayerGeoInfo, WebCassandraSvrModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/CassandraSvr/QueryCountryISOCode",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebCassandraSvrModule.QueryCountryISOCode, WebCassandraSvrModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/CassandraSvr/CssSvrUserRecharge",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebCassandraSvrModule.CssSvrUserRecharge, WebCassandraSvrModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/CassandraSvr/OldUserReceiveCompensation",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebCassandraSvrModule.OldUserReceiveCompensation, WebCassandraSvrModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/CassandraSvr/ClientCDNResourceDownloadReport",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebCassandraSvrModule.ClientCDNResourceDownloadReport, WebCassandraSvrModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/CassandraSvr/UploadUserAction",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebCassandraSvrModule.UploadUserAction, WebCassandraSvrModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/CassandraSvr/QueryUserRechargeInfo",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebCassandraSvrModule.QueryUserRechargeInfo, WebCassandraSvrModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/CassandraSvr/LilithVerifySign",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebCassandraSvrModule.LilithVerifySign, WebCassandraSvrModule.ModuleMetas)
}
