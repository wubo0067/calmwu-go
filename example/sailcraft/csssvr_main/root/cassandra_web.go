/*
 * @Author: calmwu
 * @Date: 2018-01-10 16:23:40
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-01-10 16:33:17
 * @Comment:
 */

package root

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"runtime"
	"sailcraft/base"
	"sailcraft/csssvr_main/web"
	"time"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
)

var (
	ginRouter *gin.Engine
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
	reset     = string([]byte{27, 91, 48, 109})
)

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path might contains dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastslash := bytes.LastIndex(name, slash); lastslash >= 0 {
		name = name[lastslash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}

func stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

func ginRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := stack(3)
				httprequest, _ := httputil.DumpRequest(c.Request, false)
				base.GLog.Error("[Recovery] panic recovered:\n%s\n%s\n%s%s", string(httprequest), err, stack, reset)
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}

func ginLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		c.Next()
		latency := time.Since(t)
		base.GLog.Debug("%s latency:%s", c.Request.RequestURI, latency.String())
	}
}

func init() {
	gin.SetMode(gin.DebugMode)
	ginRouter = gin.New()
	ginRouter.Use(ginLogger())
	ginRouter.Use(ginRecovery())
}

func RunWebServ(webListenIP string, webListenPort int) error {
	// 注册接口
	err := base.GinRegisterWebModule(ginRouter, web.WebCassandraSvrModule)
	if err != nil {
		base.GLog.Error("GinRegisterWebModule failed! reason[%s]", err.Error())
		return err
	}

	servAddr := fmt.Sprintf("%s:%d", webListenIP, webListenPort)
	base.GLog.Debug("CassandraSvr watch[%s]", servAddr)
	//ginRouter.Run(servAddr)
	// for gracefull restart
	endless.ListenAndServe(servAddr, ginRouter)
	return nil
}

func onHealthCheck(c *gin.Context) {
	c.Status(http.StatusOK)
}
