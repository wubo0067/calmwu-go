/*
 * @Author: calmwu 
 * @Date: 2018-10-05 22:31:07 
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-10-24 18:12:29
 */


消息转发服务，基于kafka

go build -tags static consumer.go

go build -tags static producer.go

go build -tags static compress_producer.go

git clone http://calm@10.1.41.148:8090/doyo-server-go/doyo-routersvr-go.git

bin/kafka-console-producer.sh --broker-list 10.1.41.149:9094,10.1.41.149:9092,10.1.41.149:9093 --topic test5


bin/kafka-consumer-groups.sh --all-topics --bootstrap-server 10.1.41.149:9094 --describe --group DoyoRouterSvr

go get -u -v gopkg.in/natefinch/lumberjack.v2
go get -u -v github.com/urfave/cli
go get -u -v github.com/spaolacci/murmur3
go get -u -v github.com/pquerna/ffjson/ffjson
go get -u -v github.com/monnand/dhkx
go get -u -v github.com/mitchellh/mapstructure
go get -u -v github.com/hashicorp/consul/api
go get -u -v github.com/google/go-cmp/cmp
go get -u -v github.com/emirpasic/gods/sets/hashse
go get -u -v github.com/confluentinc/confluent-kafka-go/kafka
go get -u -v gopkg.in/natefinch/lumberjack.v2
go get -u -v github.com/mozhata/merr
go get -u -v github.com/alecthomas/log4go
go get -u -v github.com/gin-gonic/gin

calmwu.kvm环境
nohup consul agent -node node0 -bind 192.168.2.200 -dev -client 192.168.2.200 &
bin/zookeeper-server-start.sh -daemon config/zookeeper.properties
bin/kafka-server-start.sh -daemon config/server.properties

./doyoreq --brokers=192.168.2.200:9092 --ip=192.168.2.200
./doyores --brokers=192.168.2.200:9092 --id=1 --ip=192.168.2.200
./doyores --brokers=192.168.2.200:9092 --id=2 --ip=192.168.2.200

./doyores --brokers=10.1.41.149:9094,10.1.41.149:9092,10.1.41.149:9093 --id=1 --ip=10.1.41.150
./doyoreq --brokers=10.1.41.149:9094,10.1.41.149:9092,10.1.41.149:9093 --ip=10.1.41.150

bin/kafka-consumer-groups.sh --bootstrap-server 10.1.41.149:9094,10.1.41.149:9092,10.1.41.149:9093 --describe --group DoyoReqApp-1

[root@localhost ~]# systemctl start supervisord.service
[root@localhost ~]# systemctl status supervisord.service

