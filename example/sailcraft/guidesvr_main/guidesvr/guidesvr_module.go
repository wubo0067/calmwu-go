/*
 * @Author: calmwu
 * @Date: 2017-12-26 15:06:00
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-06-13 15:38:18
 * @Comment:
 */

package guidesvr

import (
	"sailcraft/base"
)

type GuideSvrModule struct {
	ModuleMetas base.WebModuleMetas
}

var (
	WebGuideSvrModule *GuideSvrModule = new(GuideSvrModule)
)

func init() {
	WebGuideSvrModule.ModuleMetas = make(base.WebModuleMetas)
	base.RegisterModuleInterface("/sailcraft/api/v1/GuideSvr/LoginCheck",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebGuideSvrModule.LoginCheck, WebGuideSvrModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/GuideSvr/TuitionStepReport",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebGuideSvrModule.TuitionStepReport, WebGuideSvrModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/GuideSvr/FBInvite",
		base.HTTP_METHOD_GET,
		WebGuideSvrModule.FBInvite, WebGuideSvrModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/GuideSvr/ClientNavigate",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebGuideSvrModule.ClientNavigate, WebGuideSvrModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/GuideSvr/SetMaintainInfo",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebGuideSvrModule.SetMaintainInfo, WebGuideSvrModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/GuideSvr/ClientCDNResourceDownloadReport",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebGuideSvrModule.ClientCDNResourceDownloadReport, WebGuideSvrModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/GuideSvr/UploadUserAction",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebGuideSvrModule.UploadUserAction, WebGuideSvrModule.ModuleMetas)
}
