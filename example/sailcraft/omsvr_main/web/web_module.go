/*
 * @Author: calmwu
 * @Date: 2018-05-18 11:05:30
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-19 10:41:22
 * @Comment:
 */

package web

import "sailcraft/base"

type OMSWebModule struct {
	ModuleMetas base.WebModuleMetas
}

var (
	WebOMSModule *OMSWebModule = new(OMSWebModule)
)

func init() {
	WebOMSModule.ModuleMetas = make(base.WebModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/OMSvr/AddActiveInsts",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebOMSModule.AddActiveInsts, WebOMSModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/OMSvr/ReloadActiveInsts",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebOMSModule.ReloadActiveInsts, WebOMSModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/OMSvr/CleanAllActiveInst",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebOMSModule.CleanAllActiveInst, WebOMSModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/OMSvr/QueryRunningActiveTypes",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebOMSModule.QueryRunningActiveTypes, WebOMSModule.ModuleMetas)
}
