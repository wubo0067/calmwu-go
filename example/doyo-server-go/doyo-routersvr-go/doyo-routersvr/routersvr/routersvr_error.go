/*
 * @Author: calmwu
 * @Date: 2018-09-28 10:52:53
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-09-28 15:08:33
 */

package routersvr

import "errors"

var (
	// 转发的目的服务类型无效
	ErrServTypeInvalid = errors.New("ServType is invalid")

	// 没有有效的服务实例
	ErrNoServices = errors.New("No effective services")

	// 查询路由超时
	ErrQueryRoutinePolicyTimeOut = errors.New("Query routing policy timeout")

	// 未知错误
	ErrUnknown = errors.New("Error unknown")

	//
	ErrRoutingPolicyNotSupport = errors.New("RoutinePolicy is not support")
)
