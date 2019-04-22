/*
 * @Author: calmwu
 * @Date: 2017-11-14 14:24:32
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-04-22 15:32:18
 * @Comment:
 */
package utils

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

type TestWebModule struct {
	ModuleMetas WebItfMap
}

func (wm *TestWebModule) Init() {
	wm.ModuleMetas = make(WebItfMap)
	RegisterWebItf("/api/v1/test/Add", http.MethodGet, wm.Add1, wm.ModuleMetas)
	RegisterWebItf("/api/v2/test/Add", http.MethodGet, wm.Add2, wm.ModuleMetas)
	RegisterWebItf("/api/v2/test/Update", http.MethodPost, wm.Update, wm.ModuleMetas)
	RegisterWebItf("/api/v3/test/Delete", http.MethodDelete, wm.Delete, wm.ModuleMetas)
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

	RegisterWebItfsToGin(ginRouter, testWebModule.ModuleMetas)

	servAddr := fmt.Sprintf("%s:%d", "127.0.0.1", 8008)
	ginRouter.Run(servAddr)
}
