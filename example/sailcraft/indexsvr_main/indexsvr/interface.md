#### 请求地址

*正式环境：* `http://chkvsc.uqsoft.com`

*测试环境：* `http://192.168.1.201:505`

#### 工会名查询

*地址：* `<schema>://<host>:<port>/sailcraft/api/v1/IndexSvr/FindGuidsByName`

*Method：* `POST`

*请求内容：*

```
{
    "Version" : 1,
    "EventID" : 1,
    "TimeStamp" : xxxxxxx,
    "CsrfToken" : "abcdefg",
    "Uin" : 232232323,

    "ReqData" {
        "InterfaceName" : "FindGuidsByName",
        "Params" : {
            "GuildName" : "xxxx",
            "QueryType" : "like",       #match
            "QueryCount" : xx,          #查询条数
        }
    }
}
```

*回包内容：*

```
{
    "Version":1,
    "Timestamp":1352121016,
    "ReturnCode":0,  #0成功，非0失败

    "ResData":      
    {
        "InterfaceName" : "FindGuidsByName",
        "Params" : {
            "GuildCount" : 1,                // 搜索结果条数
            "GuildInfos" : [                   // 搜索结果
                  {
                    "ID": "1",
                    "GuildName": "公会名字",
                    "Creator": "65552",
                    "PerformId": "42120324"
                }
             ],
        }
    }
}
```


#### 玩家昵称查询

*地址：* `<schema>://<host>:<port>/sailcraft/api/v1/IndexSvr/FindUsersByName`

*Method：* `POST`


*请求内容：*

```
{
    "Version" : 1,
    "EventID" : 1,
    "TimeStamp" : xxxxxxx,
    "CsrfToken" : "abcdefg",
    "Uin" : 232232323,

    "ReqData" {
        "InterfaceName" : "FindUsersByName",
        "Params" : {
            "UserName" : "xxxx"，
            "QueryType" : "like",       #match
            "QueryCount" : xx,          #查询条数           
        }
    }
}
```

*回包内容：*

```
{
    "Version":1,
    "Timestamp":1352121016,
    "ReturnCode":0,  #0成功，非0失败

    "ResData":      
    {
        "InterfaceName" : "FindUsersByName",
        "Params" : {
            "UserCount" : xxx,  #
            "UserInfos" : [UserInfo0, UserInfo1, ...],
        }
    }

    #"UserInfo0" : {
    #    "Uin" : "xxxx",
    #    "UserName" : "xxx"
    #}
}
```

### 修改索引，修改昵称

*地址：* `<schema>://<host>:<port>/sailcraft/api/v1/IndexSvr/ModifyUserName`

*Method：* `POST`

*请求内容：*

```
{
    "Version" : 1,
    "EventID" : 1,
    "TimeStamp" : xxxxxxx,
    "CsrfToken" : "abcdefg",
    "Uin" : 232232323,

    "ReqData" {
        "InterfaceName" : "ModifyUserName",
        "Params" : {
            "UserName" : "xxxx"，
            "Uin", "xxx",
            "NewUserName" : "xxxx"          
        }
    }
}
```

#### 修改索引，工会名称 ( Obsoleted )

*地址：* `<schema>://<host>:<port>/sailcraft/api/v1/IndexSvr/ModifyGuildName`

*Method：* `POST`

*请求内容：*

```
{
    "Version" : 1,
    "EventID" : 1,
    "TimeStamp" : xxxxxxx,
    "CsrfToken" : "abcdefg",
    "Uin" : 232232323,

    "ReqData" {
        "InterfaceName" : "ModifyGuildName",
        "Params" : {
            "GuildName" : "xxxx"，
            "ID", "xxx",
            "NewGuildName" : "xxxx"         
        }
    }
}
```

#### 删除索引，工会名称、工会id

*地址：* `<schema>://<host>:<port>/sailcraft/api/v1/IndexSvr/DeleteGuildIndex`

*Method：* `POST`

*请求内容：*

```
{
    "Version" : 1,
    "EventID" : 1,
    "TimeStamp" : xxxxxxx,
    "CsrfToken" : "abcdefg",
    "Uin" : 232232323,

    "ReqData" {
        "InterfaceName" : "DeleteGuildName",
        "Params" : {
            "GuildName" : "xxxx"，
            "ID", "xxx",
            "Creator" : "xxx",
			"PerformId":"xxxx"
        }
    }
}
```


#### 添加索引，添加工会，工会名称、工会id，creator

*地址：* `<schema>://<host>:<port>/sailcraft/api/v1/IndexSvr/AddGuildIndex`

*Method：* `POST`

*请求内容：*

```
{
    "Version" : 1,
    "EventID" : 1,
    "TimeStamp" : xxxxxxx,
    "CsrfToken" : "abcdefg",
    "Uin" : 232232323,

    "ReqData" {
        "InterfaceName" : "AddGuildIndex",
        "Params" : {
            "GuildName" : "xxxx"，
            "ID", "xxx",
            "Creator" : "xxx",
			"PerformId":"xxxx"
        }
    }
}
```

#### 添加索引，添加玩家

*地址：* `<schema>://<host>:<port>/sailcraft/api/v1/IndexSvr/AddUserIndex`

*Method：* `POST`

*请求内容：*

```
{
    "Version" : 1,
    "EventID" : 1,
    "TimeStamp" : xxxxxxx,
    "CsrfToken" : "abcdefg",
    "Uin" : 232232323,

    "ReqData" {
        "InterfaceName" : "AddUserIndex",
        "Params" : {
            "UserName" : "xxxx"，
            "Uin", "xxx",
        }
    }
}
```


#### 脏字过滤

*地址：* `<schema>://<host>:<port>/sailcraft/api/v1/IndexSvr/DirtyWordFilter`

*Method：* `POST`


*请求内容：*

```
{
	"Version" : 1,
	"EventID" : 1,
	"TimeStamp" : xxxxxxx,
	"CsrfToken" : "abcdefg",
	"Uin" : 232232323,
	
	"ReqData" {
		"InterfaceName" : "DirtyWordFilter",
		"Params" : {
			"Uin" : xxx,
			"Content" : "sdasdfadfsdfsdfs"
        }
	}
}
```

*回包内容：*

```
{
    "Version":1,
    "Timestamp":1352121016,
	"ReturnCode":0,	 #0成功，非0失败

    "ResData":		
    {
		"InterfaceName" : "DirtyWordFilter",
		"Params" : {
			"Uin" : xxx,
			"FilterContent" : "*****sdsdfsd***",
            "HaveDirtyWords" : 0, // 1: 有dirty内容 0：没有dirty内容
        }
    }
}
```