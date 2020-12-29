insert into tbl_userinfo(uin, create_region, create_time, recharge_count) 
values(1, '1.1.1.1', '2011-02-03 04:05+0000', 10);

update tbl_dau set login_count=login_count+40 where date='20170503';

update tbl_userrecharge set cost=0.99, diamondtype = 1 where uin=1 and purchasetime=dateof(now());

update tbl_userrecharge set cost=0.99, diamondtype = 1 where uin=1 and purchasetime='2016-12-04 07:39:06.270473069 +0000 UTC'


UPDATE tbl_userinfo SET CreateTime='2006-01-02 15:04:05+0800', NickName="", RegisterRegion="", TotalCost=0, 
LoginTime="2006-01-02 15:04:05+0800", LogoutTime="", TradeUnionID="", VersionID=1 WHERE Uin=1

update tbl_dailyrevenue set dailytotalrechargecount = 19, dailytotalinrevenue=9.9, versionid = 1 where date='20170101' if versionid > 2;

echo "select * from ks_statisticmodule.tbl_datematchcount;"|cqlsh 139.199.62.77 9042

./dumpPreserveTables --startdate=`date -d '1 days ago' +%Y%m%d` --days=0 -c 10.135.138.179 -p ./PreserveFiles

./dumpPreserveTables --startdate=20170627 --days=5 -c 10.135.138.179 -p ./PreserveFiles

SELECT * FROM posts WHERE token(userid) > token('tom') AND token(userid) < token('bob')

SELECT firstname, lastname FROM users WHERE birth_year = 1981 AND country = 'FR' ALLOW FILTERING;

select count(*) from tbl_cdkeyinfo where cdkeybatchid=3 ALLOW FILTERING ;

0 2 * * * /srv/sandmonk/bin/dumpPreserveTables --startdate=`date -d '1 days ago' +%Y%m%d` --days=0 -c 10.161.118.88 -p /srv/sandmonk/PreserveFiles > /var/log/ca_dump.log 2>&1

 select * from tbl_userrechargerecord where time>='2018-06-22 00:00:00+0000' and time<='2018-06-23 00:00:00+0000' allow filtering;

 select * from tbl_tuitionstepstatistics2 where clientversion='1.1.8' and isocountrycode='RU';

 select * from tbl_dailyregister where date='20170906';

 copy ks_statisticmodule.tbl_dailyrevenue to '/root/revenue.csv';

 cqlsh 10.161.118.88 9042 -f statisticmodule_ks.cql

 insert into tbl_oldusercompensation(deviceid, compensationlevel, olduin, receivestatus) values('277D9F76-647E-441F-B07E-9348AB97109E', 2, 0, 0);

 下次停服时执行，修改数据表
 ALTER TABLE tbl_UserOnline ADD Platform VARCHAR;

 ALTER TABLE tbl_UserTotalRecharge ADD Platform VARCHAR;

 ALTER TABLE tbl_UserRechargeRecord ADD Platform VARCHAR; 

 update tbl_UserFinance set signininfo='{"SigninDate":20180629,"ReSigninDate":0,"MonthName":201806,"SignInCount":10}' where uin=65912;

 cqlsh 10.161.118.83 9042 -e "select * from ks_statisticmodule.tbl_tuitionstepstatistics where clientversion='1.5.4'" > tuition.csv

  cqlsh 10.161.118.83 9042 -e "select date, registercount from ks_statisticmodule.tbl_DailyRegisterCount" > DailyRegisterCount.txt

  cqlsh 10.161.118.83 9042 -e "select * from ks_statisticmodule.tbl_DailyRevenue" > DailyRevenue.txt