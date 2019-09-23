/*
 * @Author: calm.wu
 * @Date: 2019-09-23 17:37:06
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-09-23 17:54:46
 */

package utils

import (
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/DeanThompson/ginpprof"
	"github.com/gin-gonic/gin"
)

// GinRecovery middleware
func GinRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := CallStack(3)
				httprequest, _ := httputil.DumpRequest(c.Request, false)
				Errorf("[Recovery] panic recovered:\n%s\n%s\n%s", Bytes2String(httprequest), err, stack)
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}

// GinLogger middleware
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		c.Next()
		latency := time.Since(t)
		Debugf("%s latency:%s", c.Request.RequestURI, latency.String())
	}
}

// InstallPProf 安装pprof
func InstallPProf() error {
	ginForPProf := gin.New()
	ginForPProf.Use(GinLogger())
	ginForPProf.Use(GinRecovery())

	ginpprof.Wrap(ginForPProf)

	pprofHTTPSvr := &http.Server{
		Handler: ginForPProf,
	}

	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return err
	}

	// 启动监听
	go func() {
		Infof("pprof listen:%s", listener.Addr().String())
		if err := pprofHTTPSvr.Serve(listener); err != nil && err != http.ErrServerClosed {
			Fatalf("Listen %s failed. err:%s", listener.Addr().String(), err.Error())
		}
	}()
	return nil
}
