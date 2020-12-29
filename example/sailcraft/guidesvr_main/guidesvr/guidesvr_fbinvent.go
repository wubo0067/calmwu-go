/*
 * @Author: calmwu
 * @Date: 2018-02-02 10:40:27
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-02-02 10:47:22
 */

package guidesvr

import (
	"bytes"
	"html/template"
	"net/http"
	"sailcraft/base"

	"github.com/gin-gonic/gin"
)

var fbDeepLinkTemplate string = `
<html>
	<head>
		 <title>SailCraft</title>
		 <meta http-equiv="Content-Type" content="text/html; charset=utf-8">
		 <meta property="fb:app_id" content="263763767341081" />
		 <meta property="al:ios:url" content="fb263763767341081://?uin={{.UIN}}" />
		 <meta property="al:ios:app_name" content="SailCraft" />
		 <meta property="al:ios:app_store_id" content="1250365921" />
		 <meta property="al:android:package" content="com.seabattle.uq" />
		 <meta property="al:android:app_name" content="SailCraft" />
		 <meta property="al:android:url" content="com.seabattle.uq://?uin={{.UIN}}" />
		 <meta property="al:web:should_fallback" content="false" />
		 <meta property="al:web:url" content="https://www.facebook.com/games/sailcraft/?fbs=-1" />
		 <meta http-equiv="refresh" content="0;url=http://www.sailcraftonline.com/" />
 	</head>
 	<body>
 	</body>
</html>`

func (guideSvr *GuideSvrModule) FBInvite(c *gin.Context) {
	clientIP := base.GetClientAddrFromGin(c)

	if c != nil && c.Request != nil {
		base.GLog.Debug("client[%s] http.Request[%+v]", clientIP, c.Request)

		c.Request.ParseForm()
		inviteUin := c.Request.Form.Get("uin")
		base.GLog.Debug("client[%s] invide uin[%s]", clientIP, inviteUin)

		// 填写模板参数
		deepLinKTemplArgs := map[string]string{
			"UIN": inviteUin,
		}

		templ, err := template.New("fbInviteTempl").Parse(fbDeepLinkTemplate)
		if err != nil {
			base.GLog.Error("Parse fbDeepLinkTemplate failed! reason[%s]", err.Error())
			return
		}
		var deepLinkBuf bytes.Buffer
		err = templ.Execute(&deepLinkBuf, &deepLinKTemplArgs)
		if err != nil {
			base.GLog.Error("Template Execute failed! reason[%s]", err.Error())
			return
		}
		c.Data(http.StatusOK, "text/html", deepLinkBuf.Bytes())
	} else {
		base.GLog.Error("client[%s] http.Request is invalid!", clientIP)
	}
	base.GLog.Debug("client[%s] FBInviteCallback completed!", clientIP)
}
