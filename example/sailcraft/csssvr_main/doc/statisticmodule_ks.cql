// 创建keyspace
CREATE KEYSPACE IF NOT EXISTS ks_statisticmodule WITH replication = 
{'class' : 'SimpleStrategy', 'replication_factor' : 1};
// 使用keyspace
USE ks_statisticmodule;

// 日活跃统计表
CREATE TABLE IF NOT EXISTS tbl_Dau (
    Date VARCHAR,
    LoginCount COUNTER,
    PRIMARY KEY (Date)
) WITH COMMENT='DAU info table'
AND gc_grace_seconds = 604800;

// 国家活跃数
CREATE TABLE IF NOT EXISTS tbl_CountryDau (
    Date VARCHAR,
    ISOCountryCode VARCHAR,
    CountryName VARCHAR,
    LoginCount COUNTER,
    PRIMARY KEY(Date, ISOCountryCode, CountryName)
) WITH COMMENT='country dau'
AND gc_grace_seconds = 604800;

// 平台活跃数ios/android
CREATE TABLE IF NOT EXISTS tbl_PlatformDau (
    Date VARCHAR,
    Platform VARCHAR,                       // 平台系统
    LoginCount COUNTER,    
    PRIMARY KEY(Date, Platform)
) WITH COMMENT='platform dau'
AND gc_grace_seconds = 604800;

//-------------------------------------------------------------------------------------------------------

// 每日收入统计
CREATE TABLE IF NOT EXISTS tbl_DailyRevenue (
    Date VARCHAR,
    ChannelID VARCHAR,                      // CN US
    Platform VARCHAR,                       // 平台系统
    TotalRevenue FLOAT,                     // 当天的总收入
	TotalRechargeCount INT, 				// 充值次数
    VersionID BIGINT,                       // 版本信息
    PRIMARY KEY (Date, ChannelID, Platform)    
) WITH COMMENT='tbl_DailyRevenue'
AND gc_grace_seconds = 604800;

// 每日国家收入统计
CREATE TABLE IF NOT EXISTS tbl_DailyCountryRevenue (
    Date VARCHAR,
    ISOCountryCode VARCHAR,
    ChannelID VARCHAR,                      // CN US    
    Platform VARCHAR,                       // 平台系统    
    TotalRevenue FLOAT,   
	TotalRechargeCount INT, 				// 充值次数	
    VersionID BIGINT,                       // 版本信息	
    PRIMARY KEY(Date, ISOCountryCode, ChannelID, Platform) 
) WITH COMMENT='tbl_DailyCountryRevenue'
AND gc_grace_seconds = 604800;

//-------------------------------------------------------------------------------------------------------

// 日新注册数量统计表
CREATE TABLE IF NOT EXISTS tbl_DailyRegisterCount (
    Date VARCHAR,
    RegisterCount COUNTER,
    PRIMARY KEY(Date)
) WITH COMMENT='Day The number of registered users'
AND gc_grace_seconds = 604800;

// 国家日新注册统计表
CREATE TABLE IF NOT EXISTS tbl_DailyCountryRegisterCount (
    Date VARCHAR,
    ISOCountryCode VARCHAR,
    CountryName VARCHAR,
    RegisterCount COUNTER,
    PRIMARY KEY(Date, ISOCountryCode, CountryName)
) WITH COMMENT='Day The number of registered users from country'
AND gc_grace_seconds = 604800;

// 平台注册统计表
CREATE TABLE IF NOT EXISTS tbl_DailyPlatformRegisterCount (
    Date VARCHAR,
    Platform VARCHAR,                       // 平台系统
    RegisterCount COUNTER,
    PRIMARY KEY(Date, Platform)
) WITH COMMENT='Day The number of registered platform user'
AND gc_grace_seconds = 604800;

//-------------------------------------------------------------------------------------------------------

// 用户统计表
// 操作流程，先select，获得version_id，更新时用if version_id=xx，如果applied=false表明数据已经更新过了，需要
// 重新获取，保证数据的一致性。
CREATE TABLE IF NOT EXISTS tbl_UserOnline (
    Uin BIGINT,
    CreateTime TIMESTAMP,                   // 注册时间
    ISOCountryCode VARCHAR,                 // 用户注册地区
    LoginTime TIMESTAMP,                    // 登陆时间
    LogoutTime TIMESTAMP,                   // 退出时间
    MaxOnlineTime BIGINT,                   // 最大在线时长
    TotalOnlineTime BIGINT,                 // 总共在线时长
    VersionID BIGINT,
    Platform VARCHAR,                       // 平台
    PRIMARY KEY(Uin)
) WITH COMMENT='User statistics info table'
AND gc_grace_seconds = 604800;

// 用户充值表，总计
CREATE TABLE IF NOT EXISTS tbl_UserTotalRecharge (
    Uin BIGINT,
    ChannelID VARCHAR,  // CN US
    TotalCost FLOAT,
    RechargeCount INT,
    Platform VARCHAR,   // 平台    
    PRIMARY KEY(Uin, ChannelID),
) WITH COMMENT='User Total recharge table'
AND gc_grace_seconds = 604800;

// 用户充值详细记录表，每条充值记录
CREATE TABLE IF NOT EXISTS tbl_UserRechargeRecord (
    Uin BIGINT,
    ChannelID VARCHAR,                      // CN US    
    ISOCountryCode VARCHAR,
    Platform VARCHAR,   // 平台       
    Cost FLOAT,
    Time TIMESTAMP,
    PRIMARY KEY(Uin, ChannelID, Time),
) WITH COMMENT='User recharge record table'
AND gc_grace_seconds = 604800;

// 用户战斗统计
CREATE TABLE IF NOT EXISTS tbl_UserMatchInfo (
    Uin BIGINT,
    MatchCount INT,                         // 用户战斗场次
    MatchTotalTime INT,                     // 用户战斗时长, 秒
    MatchMaxDuration INT,                   // 用户一局最长时间，秒
    MatchWinCount INT,                      // 胜利次数
    MatchLostCount INT,                     // 失败次数
    MatchTieCount INT,                      // 平局次数
    MatchSurrenderCount INT,                // 投降次数
    MatchEscapeCount INT,                   // 逃跑次数
    VersionID BIGINT,
    PRIMARY KEY(Uin)
) WITH COMMENT='User Match info table'
AND gc_grace_seconds = 604800;

CREATE TABLE IF NOT EXISTS tbl_DateMatchCount (
    Date VARCHAR,
    MatchCount COUNTER,                     // 每天战斗次数
    PRIMARY KEY(Date),
) WITH COMMENT='Date match count table'
AND gc_grace_seconds = 604800;

//--------------------------------------------------------------------------------------

// 客户端新手引导步骤统计
CREATE TABLE IF NOT EXISTS tbl_TuitionStepStatistics (
    ClientVersion VARCHAR,                  // 客户端版本
    ISOCountryCode VARCHAR,
    Platform VARCHAR,                       // ios、android    
    ChannelName VARCHAR,                    // GooglePlay, AppStore    
    StepId INT,                             // 步骤id 
    Count COUNTER,                          // 计数
    PRIMARY KEY(ClientVersion, ISOCountryCode, Platform, ChannelName, StepId),
) WITH COMMENT='Tuition Step Statistics table2'
AND gc_grace_seconds = 604800;

//--------------------------------------------------------------------------------------
// 战斗录像
CREATE TABLE IF NOT EXISTS tbl_BattleVideo (
    BattleVideoID        VARCHAR,
    Time                 TIMESTAMP,
    RefCount             INT,
    Content              BLOB,
    PRIMARY KEY(BattleVideoID)
) WITH COMMENT='tbl_BattleVideo'
AND gc_grace_seconds = 604800;

// 客户端cdn资源下载结果
CREATE TABLE IF NOT EXISTS tbl_DailyClientCDNDownloadStatis (
    Date                    VARCHAR, 
    ClientVersion           VARCHAR,    // 客户端版本
    Platform                VARCHAR,    // ios、android
    ISOCountryCode          VARCHAR,    // 地域
    ResourceName            VARCHAR,    // 资源名    
    ResourceID              INT,        // 资源ID
    TotalElapseTime         COUNTER,    // 所有客户端下载总耗时
    TotalDownloadCount      COUNTER,    // 总的下载次数，上报一次加一
    TotalAttemptCount       COUNTER,    // 总的尝试次数 
    PRIMARY KEY(Date, ClientVersion, Platform, ISOCountryCode, ResourceName, ResourceID)     
) WITH COMMENT='tbl_DailyClientCDNDownloadStatis'
AND gc_grace_seconds = 604800;

// 老用户补偿表
CREATE TABLE IF NOT EXISTS tbl_OldUserCompensation (
    DeviceID            VARCHAR,
    OldUin              BIGINT,   
    CompensationLevel   INT,                // 补偿等级
    ReceiveStatus       INT,                // 领取状态
    PRIMARY KEY(DeviceID) 
) WITH COMMENT='tbl_OldUserCompensation'
AND gc_grace_seconds = 604800;

// 客户端CDN下载统计
CREATE TABLE IF NOT EXISTS tbl_DailyClientCDNDownloadStatis (
    Date                    VARCHAR, 
    ClientVersion           VARCHAR,    // 客户端版本
    Platform                VARCHAR,    // ios、android
    ISOCountryCode          VARCHAR,    // 地域
    ResourceName            VARCHAR,    // 资源名    
    ResourceID              INT,        // 资源ID
    TotalElapseTime         COUNTER,    // 所有客户端下载总耗时
    TotalDownloadCount      COUNTER,    // 总的下载次数，上报一次加一
    TotalAttemptCount       COUNTER,    // 总的尝试次数 
    PRIMARY KEY(Date, ClientVersion, Platform, ISOCountryCode, ResourceName, ResourceID)     
) WITH COMMENT='tbl_DailyClientCDNDownloadStatis'
AND gc_grace_seconds = 604800;

// 用户行为统计表
CREATE TABLE IF NOT EXISTS tbl_UserActionStatis (
    Date          VARCHAR,
    ActionName    VARCHAR,
    PerformCount  COUNTER,    // 触发次数
    PRIMARY KEY(Date, ActionName)
) WITH COMMENT='tbl_UserActionStatis'
AND gc_grace_seconds = 604800;

// 钻石消耗分类统计
CREATE TABLE IF NOT EXISTS tbl_DiamondCostTypeStatis (
    Date             VARCHAR,    
    ActionName       VARCHAR,  
    TotalDiamondCost COUNTER,       // 钻石消耗累计  
    PerformCount     COUNTER,       // 触发次数
    PRIMARY KEY(Date, ActionName)    
) WITH COMMENT='tbl_DiamondCostTypeStatis'
AND gc_grace_seconds = 604800;