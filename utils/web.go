/*
 * @Author: calmwu
 * @Date: 2017-11-14 11:11:02
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-12-08 17:28:10
 */

package utils

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
)

//WebInterfaceInfo web接口方法的描述
type WebInterfaceInfo struct {
	HttpMethodType int
	HandlerFunc    func(*gin.Context)
}

//InterfacePath 接口路径
type InterfacePath string

//WebIterfaceMap 接口集合
type WebIterfaceMap map[InterfacePath]*WebInterfaceInfo

const (
	// WEBModeuleItfs 成员变量名
	WebInterfaces = "WebInterfaces"
	// HTTPMethodGet get方法
	HTTPMethodGet = 0x0001
	// HTTPMethodPost post
	HTTPMethodPost = 0x0002
	// HTTPMethodPut put
	HTTPMethodPut = 0x0004
	// HTTPMethodDelete delete
	HTTPMethodDelete = 0x0008
)

var (
	//ErrModuleKindIsNotStruct 。。。
	ErrModuleKindIsNotStruct = errors.New("Module kind is not struct")
	//ErrModuleMetaInfosNotExist 。。。
	ErrModuleMetaInfosNotExist = errors.New("Module Interface Metainfos not exist")
	//ErrModuleMetaTypeInvalid 。。。
	ErrModuleMetaTypeInvalid = errors.New("Module meta type is not WebModuleItfInfo")

	c978WebInterfaceInfoDefault = new(WebInterfaceInfo)
	// WebInterfaceInfoType 默认类型
	WebInterfaceInfoType = reflect.TypeOf(c978WebInterfaceInfoDefault)
)

//
func GinRegisterWebModule(router *gin.Engine, webModule interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(webModule))
	t := v.Type()

	if t.Kind() == reflect.Struct {
		moduleMetaInfos := v.FieldByName(WebInterfaces)
		if !moduleMetaInfos.IsNil() {
			if moduleMetaInfos.Type().Kind() == reflect.Map {
				interfacePaths := moduleMetaInfos.MapKeys()
				for index := range interfacePaths {
					interfacePath := interfacePaths[index].String()
					fmt.Println(interfacePath)
					interfaceMetaV := moduleMetaInfos.MapIndex(interfacePaths[index])

					//fmt.Println(interfaceMetaV.Type())
					//fmt.Println(WebModuleInterfaceMetaType)

					if interfaceMetaV.Type().ConvertibleTo(WebInterfaceInfoType) {
						interfaceMeta := interfaceMetaV.Convert(WebInterfaceInfoType).Interface().(*WebModuleItfInfo)

						if (interfaceMeta.HttpMethodType & HttpMethodGet) != 0 {
							router.GET(interfacePath, interfaceMeta.HandlerFunc)
							ZLog.Info("GET apiURL[%s] registered successed!", interfacePath)
						}

						if (interfaceMeta.HttpMethodType & HttpMethodPost) != 0 {
							router.POST(interfacePath, interfaceMeta.HandlerFunc)
							ZLog.Info("POST apiURL[%s] registered successed!", interfacePath)
						}

						if (interfaceMeta.HttpMethodType & HTTPMethodPut) != 0 {
							router.PUT(interfacePath, interfaceMeta.HandlerFunc)
							ZLog.Info("PUT apiURL[%s] registered successed!", interfacePath)
						}

						if (interfaceMeta.HttpMethodType & HTTPMethodDelete) != 0 {
							router.DELETE(interfacePath, interfaceMeta.HandlerFunc)
							ZLog.Info("DELETE apiURL[%s] registered successed!", interfacePath)
							fmt.Printf("DELETE apiURL[%s] registered successed!\n", interfacePath)
						}
					} else {
						return ErrModuleMetaTypeInvalid
					}
				}
			} else {
				return ErrModuleMetaInfosNotExist
			}
		} else {
			return ErrModuleMetaInfosNotExist
		}
	} else {
		return ErrModuleKindIsNotStruct
	}
	return nil
}

func RegisterModuleInterface(interfacePath InterfacePath, httpMethodType int, handlerFunc func(*gin.Context),
	moduleMetas WebModuleItfMap) {
	if _, ok := moduleMetas[interfacePath]; !ok {
		webModuleInterfaceMeta := &WebModuleItfInfo{
			HttpMethodType: httpMethodType,
			HandlerFunc:    handlerFunc,
		}
		moduleMetas[interfacePath] = webModuleInterfaceMeta
	}
}
