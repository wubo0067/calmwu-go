/*
 * @Author: calmwu
 * @Date: 2017-11-14 14:24:32
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-11-14 15:30:44
 * @Comment:
 */
package base

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

type TestWebModule struct {
	ModuleMetas WebModuleMetas
}

func (wm *TestWebModule) Init() {
	wm.ModuleMetas = make(WebModuleMetas)
	RegisterModuleInterface("/api/v1/test/Add", HTTP_METHOD_GET, wm.Add1, wm.ModuleMetas)
	RegisterModuleInterface("/api/v2/test/Add", HTTP_METHOD_GET, wm.Add2, wm.ModuleMetas)
	RegisterModuleInterface("/api/v2/test/Update", HTTP_METHOD_POST, wm.Update, wm.ModuleMetas)
	RegisterModuleInterface("/api/v3/test/Delete", HTTP_METHOD_DELETE, wm.Delete, wm.ModuleMetas)
	fmt.Printf("ModuleMetas:%v\n", wm.ModuleMetas)
}

func (wm *TestWebModule) Add1(c *gin.Context) {
	c.String(http.StatusOK, "Add1")
}

func (wm *TestWebModule) Add2(c *gin.Context) {
	c.String(http.StatusOK, "Add2")
}

func (wm *TestWebModule) Update(c *gin.Context) {
	c.String(http.StatusOK, "Update")
}

func (wm *TestWebModule) Delete(c *gin.Context) {
	c.String(http.StatusOK, "Delete")
}

func TestWebModuleInit(t *testing.T) {
	testWebModule := new(TestWebModule)
	testWebModule.Init()
	t.Log("TestWebModuleInit test ok!")
}

// 测试地址 http://127.0.0.1:8008/api/v1/test/Add
// http://127.0.0.1:8008/api/v2/test/Add
func TestRegisterWebModule(t *testing.T) {
	gin.SetMode(gin.DebugMode)
	ginRouter := gin.Default()

	testWebModule := new(TestWebModule)
	testWebModule.Init()

	err := GinRegisterWebModule(ginRouter, testWebModule)
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log("TestRegisterWebModule test ok!")
	}

	servAddr := fmt.Sprintf("%s:%d", "127.0.0.1", 8008)
	ginRouter.Run(servAddr)
}
