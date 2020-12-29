/*
 * @Author: calmwu
 * @Date: 2017-09-20 15:08:50
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-04-28 15:35:16
 * @Comment:
 */

package indexsvr

import (
	"sailcraft/base"
)

type IndexSvrModule struct {
	ModuleMetas base.WebModuleMetas
}

var (
	WebIndexSvrModule *IndexSvrModule = new(IndexSvrModule)
)

func init() {
	WebIndexSvrModule.ModuleMetas = make(base.WebModuleMetas)
	base.RegisterModuleInterface("/sailcraft/api/v1/IndexSvr/AddGuildIndex",
		base.HTTP_METHOD_POST, WebIndexSvrModule.AddGuildIndex, WebIndexSvrModule.ModuleMetas)
	base.RegisterModuleInterface("/sailcraft/api/v1/IndexSvr/AddUserIndex",
		base.HTTP_METHOD_POST, WebIndexSvrModule.AddUserIndex, WebIndexSvrModule.ModuleMetas)
	base.RegisterModuleInterface("/sailcraft/api/v1/IndexSvr/DeleteGuildIndex",
		base.HTTP_METHOD_POST, WebIndexSvrModule.DeleteGuildIndex, WebIndexSvrModule.ModuleMetas)
	base.RegisterModuleInterface("/sailcraft/api/v1/IndexSvr/ModifyUserName",
		base.HTTP_METHOD_POST, WebIndexSvrModule.ModifyUserName, WebIndexSvrModule.ModuleMetas)
	base.RegisterModuleInterface("/sailcraft/api/v1/IndexSvr/ModifyGuildName",
		base.HTTP_METHOD_POST, WebIndexSvrModule.ModifyGuildName, WebIndexSvrModule.ModuleMetas)
	base.RegisterModuleInterface("/sailcraft/api/v1/IndexSvr/FindGuidsByName",
		base.HTTP_METHOD_POST, WebIndexSvrModule.FindGuidsByName, WebIndexSvrModule.ModuleMetas)
	base.RegisterModuleInterface("/sailcraft/api/v1/IndexSvr/FindUsersByName",
		base.HTTP_METHOD_POST, WebIndexSvrModule.FindUsersByName, WebIndexSvrModule.ModuleMetas)
	base.RegisterModuleInterface("/sailcraft/api/v1/IndexSvr/DirtyWordFilter",
		base.HTTP_METHOD_POST, WebIndexSvrModule.DirtyWordFilter, WebIndexSvrModule.ModuleMetas)
}
