1. Consul

```
机器：192.168.68.229
启动命令：nohup consul agent -node node0 -bind 192.168.68.229 -dev -client 192.168.68.229 &
查看：http://192.168.68.229:8500/ui/dc1/services
```
2. ZooKeeper
```
机器：192.168.68.229

wget http://ftp.cuhk.edu.hk/pub/packages/apache.org/zookeeper/current/zookeeper-3.4.12.tar.gz

doyo_1.cfg配置：
tickTime=2000
initLimit=10
syncLimit=5
dataDir=/data/calmwu/3rd/zookeeper-3.4.12/doyo_zk1/datadir
dataLogDir=/data/calmwu/3rd/zookeeper-3.4.12/doyo_zk1/logdir
clientPort=7181
server.1=127.0.0.1:4888:5888
server.2=127.0.0.1:4889:5889
server.3=127.0.0.1:4890:5890

doyo_2.cfg配置：
tickTime=2000
initLimit=10
syncLimit=5
dataDir=/data/calmwu/3rd/zookeeper-3.4.12/doyo_zk2/datadir
dataLogDir=/data/calmwu/3rd/zookeeper-3.4.12/doyo_zk2/logdir
clientPort=7182
server.1=127.0.0.1:4888:5888
server.2=127.0.0.1:4889:5889
server.3=127.0.0.1:4890:5890

doyo_3.cfg配置：
tickTime=2000
initLimit=10
syncLimit=5
dataDir=/data/calmwu/3rd/zookeeper-3.4.12/doyo_zk3/datadir
dataLogDir=/data/calmwu/3rd/zookeeper-3.4.12/doyo_zk3/logdir
clientPort=7183
server.1=127.0.0.1:4888:5888
server.2=127.0.0.1:4889:5889
server.3=127.0.0.1:4890:5890

在每个实例的datadir目录下执行
echo 1 > myid
echo 2 > myid
echo 3 > myid

./bin/zkServer.sh start ./conf/doyo_1.cfg 
./bin/zkServer.sh start ./conf/doyo_2.cfg 
./bin/zkServer.sh start ./conf/doyo_3.cfg 
```

3. kafka
```
机器：192.168.68.230，
监听端口分别为：listeners=PLAINTEXT://192.168.68.230:9092
              listeners=PLAINTEXT://192.168.68.230:9093
              listeners=PLAINTEXT://192.168.68.230:9094

下载安装包：
wget https://www.apache.org/dyn/closer.cgi?path=/kafka/2.0.0/kafka_2.11-2.0.0.tgz
tar -zxf kafka_2.11-2.0.0.tgz

zookeeper.connect=192.168.68.229:7181,192.168.68.229:7182,192.168.68.229:7183

为每个实例创建log目录
mkdir -p /data/calmwu/3rd/kafka_2.11-2.0.0/kfk1-logs
mkdir -p /data/calmwu/3rd/kafka_2.11-2.0.0/kfk2-logs
mkdir -p /data/calmwu/3rd/kafka_2.11-2.0.0/kfk3-logs

生成3个配置文件
-rw-r--r-- 1 calmwu users 6938 11月  9 11:02 server-1.properties
-rw-r--r-- 1 calmwu users 6938 11月  9 11:03 server-2.properties
-rw-r--r-- 1 calmwu users 6938 11月  9 11:03 server-3.properties

启动
bin/kafka-server-start.sh -daemon config/server-1.properties
bin/kafka-server-start.sh -daemon config/server-2.properties
bin/kafka-server-start.sh -daemon config/server-3.properties

创建topic
bin/kafka-topics.sh --create --zookeeper 192.168.68.229:7181,192.168.68.229:7182,1192.168.68.229:7183 --replication-factor 3 --partitions 3 --topic test5

查看
bin/kafka-topics.sh --describe --zookeeper 192.168.68.229:7181
```

