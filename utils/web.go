/*
 * @Author: calmwu
 * @Date: 2017-11-14 11:11:02
 * @Last Modified by: calmwu
 * @Last Modified time: 2020-08-15 20:06:38
 */

package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

//WebInterfaceInfo web接口方法的描述
type WebInterfaceInfo struct {
	HTTPMethodType string
	HandlerFunc    gin.HandlerFunc
}

//WebItfMap 接口集合
type WebItfMap map[string]*WebInterfaceInfo

// RegisterWebItfsToGin 注册到gin
func RegisterWebItfsToGin(router *gin.Engine, webItfMap WebItfMap) {
	var ginHandlerFunc func(string, ...gin.HandlerFunc) gin.IRoutes

	for webItfPath, webItfInfo := range webItfMap {
		switch webItfInfo.HTTPMethodType {
		case http.MethodGet:
			ginHandlerFunc = router.GET
		case http.MethodPost:
			ginHandlerFunc = router.POST
		case http.MethodPut:
			ginHandlerFunc = router.PUT
		case http.MethodDelete:
			ginHandlerFunc = router.DELETE
		default:
			ZLog.Errorf("ItfPath:%s MethodType:%s not support!", webItfPath, webItfInfo.HTTPMethodType)
			continue
		}

		ginHandlerFunc(webItfPath, webItfInfo.HandlerFunc)
		ZLog.Infof("Register ItfPath:%s MethodType:%s to GinRouter", webItfPath, webItfInfo.HTTPMethodType)
	}
}

// RegisterWebItf 接口注册
func RegisterWebItf(webItfPath, httpMethodType string, handlerFunc gin.HandlerFunc, webItfMap WebItfMap) {
	if _, ok := webItfMap[webItfPath]; !ok {
		webItfInfo := &WebInterfaceInfo{
			HTTPMethodType: httpMethodType,
			HandlerFunc:    handlerFunc,
		}
		webItfMap[webItfPath] = webItfInfo
	}
}
