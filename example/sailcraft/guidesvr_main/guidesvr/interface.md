## 说明
***
联调环境：
```
stage  http://123.59.40.19:400
dev-8885 http://192.168.1.201:805
dev-8889 http://192.168.1.201:809

正式环境，需要分开、域名是独立的
checklogin http://chksvrsc2.uqsoft.com:800
navigate http://navigationsc2.uqsoft.com:801
```

### 接口
***
1. LoginCheck
```json
  InterfaceName: LoginCheck
  Url: /sailcraft/api/v1/GuideSvr/LoginCheck
  Request:
{
    "Version":1,
    "EventId":121,          # 事件ID
    "Timestamp":1352121016,
    "CsrfToken":"02cf14994a3be74301657dbcd9c0a189",    # 固定的一个校验码
    "ChannelUID" "xxx",    #渠道id
    "Uin":123456,         # 玩家的Id，没有填0
    "ReqData":
    {
        "InterfaceName":"LoginCheck",
        "Params":{
            "ClientVersion":"x.x.x",        # 当前版本号
            "PlatformName":"Android",       # Ios，平台
            "ChannelName":"GooglePlay",     # AppStore，渠道     
        }
    }
}
  Response:
{
    "Version":1,
    "Timestamp":1352121016,
    "ReturnCode":0, 
    # -1: 错误。
    # 0：服务器正常，可以正常登录。
    # 1：禁止登陆，黑名单用户。
    # 2：版本不匹配、需要更新。
    # 3：服务器在维护中

    "ResData":      #返回数据，返回值不同填写数据不同
    {
        "InterfaceName":"LoginCheck",
        "Params" : {
            #RetuenCode=-1
            FailureReason: "xxxxxx"

            #ReturnCode=0:
            ServerIPs : ["x.x.x.x", "x.x.x.x", "x.x.x.x"],
            Port: 7045,
            ClientInternetIP : "x.x.x.x",

            #ReturnCode=2
            NewVersion : "x.x.x",
            ChannelName : "GooglePlay",     # AppStore，渠道
            UpdateUrl : "更新跳转地址",

            #ReturnCode=3
            Bulletin : "停服维护公告信息",
            RemainingSeconds : xxx,         # 开服预计剩余秒数 UTC时间
        }
    }
}
```

2. 新手引导上报
```json
  InterfaceName: TuitionStepReport
  Url: /sailcraft/api/v1/GuideSvr/TuitionStepReport
{
    "Version":1,
    "EventId":121,          # 事件ID
    "Timestamp":1352121016,
    "CsrfToken":"02cf14994a3be74301657dbcd9c0a189",    # 固定的一个校验码
    "ChannelUID" "xxx",    #渠道id
    "Uin":123456,         # 玩家的Id，没有填0
    "ReqData":
    {
        "InterfaceName":"TuitionStepReport",
        "Params":{
            "ClientVersion":"x.x.x",        # 当前版本号
            "StepId":1,                     # 统计的步骤id 1-54步 
            "PlatformName":"Android",       # Ios，平台
            "ChannelName":"GooglePlay",     # AppStore，渠道         
        }
    }
}
```

3. 客户端登录引导，根据客户端版本号引导客户端进入正式服、审核服
```json
  InterfaceName: ClientNavigate
  Url: /sailcraft/api/v1/GuideSvr/ClientNavigate
  Request:
{
    "Version":1,
    "EventId":121,          # 事件ID
    "Timestamp":1352121016,
    "CsrfToken":"02cf14994a3be74301657dbcd9c0a189",    # 固定的一个校验码
    "ChannelUID" "xxx",    #渠道id
    "Uin":123456,         # 玩家的Id，没有填0
    "ReqData":
    {
        "InterfaceName":"ClientNavigate",
        "Params":{
            "ClientVersion":"x.x.x",        # 当前版本号
            "PlatformName":"Android",       # Ios，平台
            "ChannelName":"GooglePlay",     # AppStore，渠道     
        }
    }
}
  Response:
{
    "Version":1,
    "Timestamp":1352121016,
    "ReturnCode":0, 
    # -1: 错误。
    # 0：ok

    "ResData":      #返回数据，返回值不同填写数据不同
    {
        "InterfaceName":"ClientNavigate",
        "Params" : {
            "ULC" : "chksvrsc2.uqsoft.com:800",     // UrlLoginCheckd地址，example: https://chksvrsc2.uqsoft.com:802
            "UPS" : "proxysc2.uqsoft.com:6483",     // ProxySvr地址，example: proxysc2.uqsoft.com:6483
            "USS" : "sdksvrsc2.uqsoft.com:80",      // SdkSvrd地址，example: https://sdksvrsc2.uqsoft.com:80
        }
    }
}
```

4. 更新停服标志，白名单 ***特别说明，在现网用域名的方式进行访问 SailCraft-GuideSvr.service.consul:8000 ***
```json
  InterfaceName: SetMaintainInfo
  Url: /sailcraft/api/v1/GuideSvr/SetMaintainInfo
  Request:
{
    "Version":1,
    "EventId":121,          # 事件ID
    "Timestamp":1352121016,
    "CsrfToken":"02cf14994a3be74301657dbcd9c0a189",    # 固定的一个校验码
    "ChannelUID" "xxx",    #渠道id
    "Uin":123456,         # 玩家的Id，没有填0
    "ReqData":
    {
        "InterfaceName":"SetMaintainInfo",
		"Params" : {
			"gm_tool_game_maintain_key" : {
				"white_flag" : 1,
				"game_maintain_flag" : 0,
				"white_list" : ["xxx", "xxx", "xxx"],
				"main_dead_line" : xxxx, 
			},
		}
    }
}
  Response:
{
    "Version":1,
    "Timestamp":1352121016,
    "ReturnCode":0, 
    # -1: 错误。
    # 0：ok
}
```

5. 客户端CDN资源下载上报
```json
  InterfaceName: ClientCDNResourceDownloadReport
  Url: /sailcraft/api/v1/GuideSvr/ClientCDNResourceDownloadReport
  Request:
{
    "Version":1,
    "EventId":121,          # 事件ID
    "Timestamp":1352121016,
    "CsrfToken":"02cf14994a3be74301657dbcd9c0a189",    # 固定的一个校验码
    "ChannelUID" "xxx",    #渠道id
    "Uin":123456,         # 玩家的Id，没有填0
    "ReqData":
    {
        "InterfaceName":"ClientCDNResourceDownloadReport",
        "Params":{
			"ClientVersion":"x.x.x",    			# 当前版本号            string
			"ResourceName":"assertbundle-xxx",	    # 资源名字				string
			"ResourceID":xx,						# 资源的下载索引		  int
			"ElapseTime":xx,						# 下载耗时				int
			"AttemptCount":xxx,						# 尝试次数				int
			"PlatformName":"Android",	    		# Ios，平台
			"ChannelName":"GooglePlay",				# AppStore，渠道			
        }
    }
}
```

6. 用户行为统计
```
json
  InterfaceName: UploadUserAction
  Url: /sailcraft/api/v1/GuideSvr/UploadUserAction
  Request:
{
    "Version":1,
    "EventId":121,          # 事件ID
    "Timestamp":1352121016,
    "CsrfToken":"02cf14994a3be74301657dbcd9c0a189",    # 固定的一个校验码
    "ChannelUID" "xxx",    #渠道id
    "Uin":123456,         # 玩家的Id，没有填0
    "ReqData":
    {
        "InterfaceName":"UploadUserAction",
        "Params":{
			"Uin":65555,    			   # 用户id
			"ActionName": "Client_XXXX",   # 行为名，以Client_为前缀，后缀由调用方根据运营需求设定，接口负责对事件进行统计。现在运营需求：点开月卡界面、点超值礼包向右切换箭头、点开首充面板、点开充值面板
        }
    }
}
```