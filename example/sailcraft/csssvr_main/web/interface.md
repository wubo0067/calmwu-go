## 接口说明
***

1. 上传战斗视频  
   URL: http://IP:9000/sailcraft/api/v1/CassandraSvr/UploadBattleVideo  
   InterfaceName: UploadBattleVideo  


2. 查询战斗视频  
    URL: http://IP:9000/sailcraft/api/v1/CassandraSvr/GetBattleVideo  
    InterfaceName: GetBattleVideo  


3. 删除战斗视频  
   URL: http://IP:9000/sailcraft/api/v1/CassandraSvr/DeleteBattleVideo    
   InterfaceName: DeleteBattleVideo


4. 根据客户端IP查询国家代码  
    URL: http://IP:9000/sailcraft/api/v1/CassandraSvr/QueryCountryISOCode  
    InterFaceName: QueryCountryISOCode
    ```json
    Request:
    "Param" : {
        "Uin" : 3232,
        "ClientIP" : x.x.x.x,
    }  
    Response:  
    "Param" : {  
        "Uin" : 3232,
        "ClientIP" : x.x.x.x,
        "CountryISOCode" : "xxx",
        "CountryName" : "XXX",
    }
    ```

5. 用户登录  
    URL: http://IP:9000/sailcraft/api/v1/CassandraSvr/CssSvrUserLogin  
    InterFaceName: QueryCountryISOCode
    ```json
    Request:
    "Param" : {
        "Uin" : 3232,
        "ClientInternetIP" : x.x.x.x,
        "Platform" : IOS, //ANDROID
    }  
    ```

5. 用户退出 
    URL: http://IP:9000/sailcraft/api/v1/CassandraSvr/CssSvrUserLogout  
    InterFaceName: QueryCountryISOCode
    ```json
    Request:
    "Param" : {
        "Uin" : 3232,
        "ClientInternetIP" : x.x.x.x,
        "Platform" : IOS, //ANDROID
    }  
    ```

6. 老用户领取补偿
    URL: http://IP:9000/sailcraft/api/v1/CassandraSvr/OldUserReceiveCompensation
    InterFaceName: OldUserReceiveCompensation
    ```json
    Request:
    "Param" : {
        "DeviceID" : "3232",
    }  
    Response:  
    "Param" : {  
        "DeviceID" : "3232",
        "Result" : 0, // 0: 领取成功，-1：领取失败，已经领取过
        "Level" : 1, // 1,2,3
    }
    ```    

7. 用户行为上报
    URL: http://IP:9000/sailcraft/api/v1/CassandraSvr/UploadUserAction
    InterFaceName: UploadUserAction
    ```json
    Request:
    "Param" : {
        "Uin" : "3232",
        "ActionName" : "Server_XXXX", //以Server_为前缀，后缀由调用方根据运营需求设定，接口负责对事件进行统计。现在运营需求：累计消费每个档次点击，各种钻石消耗行为
        "DiamondCostCount" : 1212, // 如果是钻石消耗行为，这里填消耗的数量
    }  
    ```        

8. 查询用户充值信息
    URL: http://IP:9000/sailcraft/api/v1/CassandraSvr/QueryUserRechargeInfo
    InterFaceName: QueryUserRechargeInfo
    ```json
    Request:
    "Param" : {
        "Uin" : 3232,
    }
    Response:  
    "Param" : {  
        "Uin" : 3232,
        "MaxRechargeAmount" : 2323.23, // 
        "TotalRechargeAmount" : 2323232323.22, //
        "TotalRechargeCount" : 10,
    }     
    ```    