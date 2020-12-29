CREATE DATABASE IF NOT EXISTS omsdb;
USE omsdb;

DROP TABLE IF EXISTS tbl_ActiveInstControl;

CREATE TABLE IF NOT EXISTS tbl_ActiveInstControl (
 Id INT AUTO_INCREMENT,
 ZoneID INT NOT NULL ,
 ActiveType INT, 
 ActiveID INT,
 StartTime VARCHAR(256), 
 DurationMinutes INT,
 ChannelName VARCHAR(8),
 TimeZone VARCHAR(256),
 PerformState INT,
 GroupID int, 
 INDEX(ZoneID),
 INDEX(ActiveType),
 INDEX(ActiveID),
 INDEX(PerformState), 
 PRIMARY KEY ( Id )) ENGINE=InnoDB DEFAULT CHARSET=utf8;