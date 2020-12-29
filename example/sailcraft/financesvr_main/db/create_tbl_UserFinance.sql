CREATE DATABASE IF NOT EXISTS user_finance;
USE user_finance;

DROP TABLE IF EXISTS tbl_UserFinance;
DROP TABLE IF EXISTS tbl_NewPlayerLoginBenefits;
DROP TABLE IF EXISTS tbl_PlayerActive;
DROP TABLE IF EXISTS tbl_PlayerCDKeyExchange;

CREATE TABLE IF NOT EXISTS tbl_UserFinance (
 Uin BIGINT UNSIGNED NOT NULL ,
 ZoneID INT NOT NULL ,
 TimeZone VARCHAR(64), 
 FirstRecharge VARCHAR(256),
 VipInfo VARCHAR(256), 
 ShopFirstPurchaseInfo VARCHAR(512),
 PlayerRefreshShopDailyInfo TEXT,
 SignInInfo TEXT, 
 INDEX(ZoneID),
 PRIMARY KEY(Uin) ) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ALTER TABLE tbl_UserFinance CHANGE RechargeShopFirstPurchaseInfo ShopFirstPurchaseInfo VARCHAR(512);
-- ALTER TABLE tbl_UserFinance MODIFY ShopFirstPurchaseInfo VARCHAR(512);
--  alter table `tbl_UserFinance` add column WeeklySignInInfo TEXT AFTER `PlayerRefreshShopDailyInfo`;
--  alter table `tbl_UserFinance` add column RechargeShopFirstPurchaseInfo VARCHAR(128) AFTER `ConsumptionCardEndtime`;

CREATE TABLE IF NOT EXISTS tbl_NewPlayerLoginBenefits (
 Uin BIGINT UNSIGNED NOT NULL,
 ZoneID INT NOT NULL,
 TimeZone VARCHAR(64),  
 CreateDate INT NOT NULL,
 LoginDays INT NOT NULL,
 LastLoginDate INT NOT NULL,
 ReceiveAwardTags VARCHAR(32),
 ReceiveAwardCount INT NOT NULL,
 IsCompleted INT NOT NULL,
 INDEX(ZoneID),
PRIMARY KEY(Uin) ) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- alter table `tbl_NewPlayerLoginBenefits` add column TimeZone VARCHAR(64) AFTER `ZoneID`;

CREATE TABLE IF NOT EXISTS tbl_PlayerActive (
 Uin BIGINT UNSIGNED NOT NULL,
 ZoneID INT NOT NULL,
 ActiveType INT NOT NULL,
 ActiveID INT NOT NULL,
 ChannelID VARCHAR(32),
 AccumulateCount INT NOT NULL,
 ReceiveCount INT,
 ActiveStartTime DateTime NOT NULL,
 ActiveEndTime DateTime NOT NULL,  
 ActiveResetTime DateTime NOT NULL,
 TaskType VARCHAR(32),
 INDEX(ZoneID),
 INDEX(TaskType),
PRIMARY KEY(Uin, ActiveType, ActiveID) ) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS tbl_PlayerCDKeyExchange (
 Uin BIGINT UNSIGNED NOT NULL,
 ZoneID INT NOT NULL,
 KHLst TEXT,
PRIMARY KEY(Uin) ) ENGINE=InnoDB DEFAULT CHARSET=utf8;