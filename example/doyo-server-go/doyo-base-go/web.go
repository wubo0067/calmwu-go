/*
 * @Author: calmwu
 * @Date: 2017-11-14 11:11:02
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-12-08 17:28:10
 */

package base

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
)

// web接口方法的描述
type WebModuleInterfaceMeta struct {
	HttpMethodType int
	HandlerFunc    func(*gin.Context)
}

// web模块的描述集合
type InterfacePath string
type WebModuleMetas map[InterfacePath]*WebModuleInterfaceMeta

const (
	WEBMODULE_METAS    = "ModuleMetas"
	HTTP_METHOD_GET    = 0x0001
	HTTP_METHOD_POST   = 0x0002
	HTTP_METHOD_PUT    = 0x0004
	HTTP_METHOD_DELETE = 0x0008
)

var (
	ErrModuleKindIsNotStruct   = errors.New("Module kind is not struct")
	ErrModuleMetaInfosNotExist = errors.New("Module Interface Metainfos not exist")
	ErrModuleMetaTypeInvalid   = errors.New("Module meta type is not WebModuleInterfaceMeta")

	c_978_WebModuleInterfaceMeta_Default = new(WebModuleInterfaceMeta)
	WebModuleInterfaceMetaType           = reflect.TypeOf(c_978_WebModuleInterfaceMeta_Default)
)

//
func GinRegisterWebModule(router *gin.Engine, webModule interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(webModule))
	t := v.Type()

	if t.Kind() == reflect.Struct {
		moduleMetaInfos := v.FieldByName(WEBMODULE_METAS)
		if !moduleMetaInfos.IsNil() {
			if moduleMetaInfos.Type().Kind() == reflect.Map {
				interfacePaths := moduleMetaInfos.MapKeys()
				for index := range interfacePaths {
					interfacePath := interfacePaths[index].String()
					fmt.Println(interfacePath)
					interfaceMetaV := moduleMetaInfos.MapIndex(interfacePaths[index])

					//fmt.Println(interfaceMetaV.Type())
					//fmt.Println(WebModuleInterfaceMetaType)

					if interfaceMetaV.Type().ConvertibleTo(WebModuleInterfaceMetaType) {
						interfaceMeta := interfaceMetaV.Convert(WebModuleInterfaceMetaType).Interface().(*WebModuleInterfaceMeta)

						if (interfaceMeta.HttpMethodType & HTTP_METHOD_GET) != 0 {
							router.GET(interfacePath, interfaceMeta.HandlerFunc)
							ZLog.Info("GET apiURL[%s] registered successed!", interfacePath)
						}

						if (interfaceMeta.HttpMethodType & HTTP_METHOD_POST) != 0 {
							router.POST(interfacePath, interfaceMeta.HandlerFunc)
							ZLog.Info("POST apiURL[%s] registered successed!", interfacePath)
						}

						if (interfaceMeta.HttpMethodType & HTTP_METHOD_PUT) != 0 {
							router.PUT(interfacePath, interfaceMeta.HandlerFunc)
							ZLog.Info("PUT apiURL[%s] registered successed!", interfacePath)
						}

						if (interfaceMeta.HttpMethodType & HTTP_METHOD_DELETE) != 0 {
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
	moduleMetas WebModuleMetas) {
	if _, ok := moduleMetas[interfacePath]; !ok {
		webModuleInterfaceMeta := &WebModuleInterfaceMeta{
			HttpMethodType: httpMethodType,
			HandlerFunc:    handlerFunc,
		}
		moduleMetas[interfacePath] = webModuleInterfaceMeta
	}
}
