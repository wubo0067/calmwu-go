/*
 * @Author: calmwu
 * @Date: 2018-02-01 17:56:26
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-18 11:06:16
 * @Comment:
 */

package web

import "sailcraft/base"

type FinanceWebModule struct {
	ModuleMetas base.WebModuleMetas
}

var (
	WebFinanceModule *FinanceWebModule = new(FinanceWebModule)
)

func init() {
	WebFinanceModule.ModuleMetas = make(base.WebModuleMetas)
	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/QueryRechargeCommodities",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.QueryRechargeCommodities, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/DeliveryRechargeCommodity",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.DeliveryRechargeCommodity, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/RefreshRechargeCommodities",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.RefreshRechargeCommodities, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/NewFinanceUser",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.NewFinanceUser, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/QueryUserVIPType",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.QueryUserVIPType, WebFinanceModule.ModuleMetas)

	//-------------------------------------------------------------------------------------

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/QueryResourceCommdities",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.QueryResourceCommdities, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/RefreshResourceShopConfig",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.RefreshResourceShopConfig, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/BuyResourceCommodity",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.BuyResourceCommodity, WebFinanceModule.ModuleMetas)

	//-------------------------------------------------------------------------------------

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/QueryCardPackCommdities",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.QueryCardPackCommdities, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/RefreshCardPackShopConfig",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.RefreshCardPackShopConfig, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/BuyCardPackCommodity",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.BuyCardPackCommodity, WebFinanceModule.ModuleMetas)

	//-------------------------------------------------------------------------------------

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/GetRefreshShopCommodities",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.GetRefreshShopCommodities, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/UpdateRefreshShopConfig",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.UpdateRefreshShopConfig, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/UpdateRefreshShopCommodityPool",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.UpdateRefreshShopCommodityPool, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/BuyRefreshShopCommodity",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.BuyRefreshShopCommodity, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/CheckRefreshShopManualRefresh",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.CheckRefreshShopManualRefresh, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/GetRefreshShopCommodityCost",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.GetRefreshShopCommodityCost, WebFinanceModule.ModuleMetas)

	//-------------------------------------------------------------------------------------

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/GetUserFinanceBusinessRedLights",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.GetUserFinanceBusinessRedLights, WebFinanceModule.ModuleMetas)

	//-------------------------------------------------------------------------------------

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/GMUpdateMonthlySigninConfigInfo",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.GMUpdateMonthlySigninConfigInfo, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/GetMonthlySigninInfo",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.GetMonthlySigninInfo, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/PlayerSignIn",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.PlayerSignIn, WebFinanceModule.ModuleMetas)

	//-------------------------------------------------------------------------------------

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/GMConfigVIPPrivilege",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.GMConfigVIPPrivilege, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/GetPlayerVIPInfo",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.GetPlayerVIPInfo, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/VIPPlayerCollectPrize",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.VIPPlayerCollectPrize, WebFinanceModule.ModuleMetas)

	//-------------------------------------------------------------------------------------

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/GMConfigNewPlayerLoginBenefit",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.GMConfigNewPlayerLoginBenefit, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/GetNewPlayerLoginBenefitInfo",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.GetNewPlayerLoginBenefitInfo, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/ReceiveLoginBenefit",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.ReceiveLoginBenefit, WebFinanceModule.ModuleMetas)

	//-------------------------------------------------------------------------------------

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/OpenActive",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.OpenActive, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/CloseActive",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.CloseActive, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/GMConfigSuperGiftActive",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.GMConfigSuperGiftActive, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/GMConfigMissionActive",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.GMConfigMissionActive, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/GMConfigExchangeActive",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.GMConfigExchangeActive, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/GMConfigCDKeyExchangeActive",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.GMConfigCDKeyExchangeActive, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/GetPlayerActive",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.GetPlayerActive, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/ActiveAccumulateParameterNtf",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.ActiveAccumulateParameterNtf, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/PlayerActiveReceive",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.PlayerActiveReceive, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/GetActiveExchangeCost",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.GetActiveExchangeCost, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/CheckActiveConfig",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.CheckActiveConfig, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/PlayerExchangeCDKey",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.PlayerExchangeCDKey, WebFinanceModule.ModuleMetas)

	//-------------------------------------------------------------------------------------

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/GMConfigFirstRecharge",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.GMConfigFirstRecharge, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/GetFirstRechargeActive",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.GetFirstRechargeActive, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/ReceiveFirstRechargeReward",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.ReceiveFirstRechargeReward, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/CheckPlayerActiveIsCompleted",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.CheckPlayerActiveIsCompleted, WebFinanceModule.ModuleMetas)

	base.RegisterModuleInterface("/sailcraft/api/v1/FinanceSvr/QueryRechargeCommodityPrices",
		base.HTTP_METHOD_POST|base.HTTP_METHOD_PUT,
		WebFinanceModule.QueryRechargeCommodityPrices, WebFinanceModule.ModuleMetas)
}
