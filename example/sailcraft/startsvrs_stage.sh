#!/bin/bash

cd csssvr_main/bin/stage
./start.sh
sleep 1
svr_count="`ps -ef|grep csssvr_main|grep -v 'grep'|wc -l`"
if [ $svr_count -gt 0 ]
then
    echo "csssvr_main is start!  `date`"|tee -a /var/log/message
else
    echo "csssvr_main is start failed!  `date`"|tee -a /var/log/message
fi
cd -

cd fleetsvr_main/bin/stage
./startfleetsvr.sh
sleep 1
svr_count="`ps -ef|grep fleetsvr_main|grep -v 'grep'|wc -l`"
if [ $svr_count -gt 0 ]
then
    echo "fleetsvr_main is start!  `date`"|tee -a /var/log/message
else
    echo "fleetsvr_main is start failed!  `date`"|tee -a /var/log/message
fi
cd -

cd indexsvr_main/bin/stage
./start.sh
sleep 1
svr_count="`ps -ef|grep indexsvr_main|grep -v 'grep'|wc -l`"
if [ $svr_count -gt 0 ]
then
    echo "indexsvr_main is start!  `date`"|tee -a /var/log/message
else
    echo "indexsvr_main is start failed!  `date`"|tee -a /var/log/message
fi
cd -

cd financesvr_main/bin/stage
./start.sh
sleep 1
svr_count="`ps -ef|grep financesvr_main|grep -v 'grep'|wc -l`"
if [ $svr_count -gt 0 ]
then
    echo "financesvr_main is start!  `date`"|tee -a /var/log/message
else
    echo "financesvr_main is start failed!  `date`"|tee -a /var/log/message
fi
cd -

cd guidesvr_main/bin/stage
./start.sh
sleep 1
svr_count="`ps -ef|grep guidesvr_main|grep -v 'grep'|wc -l`"
if [ $svr_count -gt 0 ]
then
    echo "guidesvr_main is start!  `date`"|tee -a /var/log/message
else
    echo "guidesvr_main is start failed!  `date`"|tee -a /var/log/message
fi
cd -

cd logsvr_main/bin/stage
./start.sh
sleep 1
svr_count="`ps -ef|grep logsvr_main|grep -v 'grep'|wc -l`"
if [ $svr_count -gt 0 ]
then
    echo "logsvr_main is start!  `date`"|tee -a /var/log/message
else
    echo "logsvr_main is start failed!  `date`"|tee -a /var/log/message
fi
cd -

#cd omsvr_main/bin/stage
#./start.sh
#svr_count="`ps -ef|grep omsvr_main|grep -v 'grep'|wc -l`"
#if [ $svr_count -gt 0 ]
#then
#    echo "omsvr_main is start!  `date`"|tee -a /var/log/message
#else
#    echo "omsvr_main is start failed!  `date`"|tee -a /var/log/message
#fi
#cd -