1. go build -o shopConfig.exe main.go update_resourceshop.go update_cardpackshop.go update_refreshshop.go update_rechargeshop.go update_monthsignin.go update_vipconfig.go update_newplayerbenefit.go update_activesupergift.go update_activemission.go update_activeexchange.go update_activecdkey.go update_firstrecharge.go

2. .\shopConfig.exe --type=resource --configpath=../doc/Shop --zoneid=1 --version=11

3. .\shopConfig.exe --type=cardpack --configpath=../doc/Shop --zoneid=1 --version=11

4. .\shopConfig.exe --type=refresh --configpath=../doc/Shop --zoneid=1 --version=12

5. .\shopConfig.exe --type=recharge --configpath=../doc/Shop --zoneid=1 --version=12

6. .\shopConfig.exe --type=signin --configpath=../doc/Shop --zoneid=1 --version=1

7. .\shopConfig.exe --type=vip --configpath=../doc/Shop --zoneid=1 --version=1

8. .\shopConfig.exe --type=newplayerbenefit --configpath=../doc/Shop --zoneid=1 --version=1

9. .\shopConfig.exe --type=supergift --configpath=../doc/Shop --zoneid=1 --version=1

9. .\shopConfig.exe --type=mission --configpath=../doc/Shop --zoneid=1 --version=1

10. .\shopConfig.exe --type=exchange --configpath=../doc/Shop --zoneid=1 --version=1

11. .\shopConfig.exe --type=cdkey --configpath=../doc/Shop --zoneid=1 --version=1

12. .\shopConfig.exe --type=firstrecharge --configpath=../doc/Shop --zoneid=1 --version=1

12. ./shopConfig.exe --type=resource --configpath=../doc/Shop --zoneid=1 --version=1 --svrip=192.168.1.201:4000