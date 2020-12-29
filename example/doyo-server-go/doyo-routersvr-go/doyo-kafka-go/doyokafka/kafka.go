/*
 * @Author: calmwu
 * @Date: 2018-09-15 14:28:09
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-09-30 10:51:06
 */

package doyokafka

import (
	"fmt"
	"strings"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
)

const (
	KREADCHAN_SIZE  = 1024
	KWRITECHAN_SIZE = 1024
)

type KafkaModule struct {
	brokers string
	topics  []string
	groupid string

	producer *kafka.Producer
	consumer *kafka.Consumer

	kExitChan chan struct{}
	kPullChan chan DoyoKafkaData // 读管道，从kafka收到的数据放到里面
	kPushChan chan DoyoKafkaData // 写管道，接收从业务过来的数据

	logger   *zap.SugaredLogger
	exitWait sync.WaitGroup
}

func InitModule(brokers string, topics []string, groupid string, logger *zap.SugaredLogger) (*KafkaModule, error) {
	var err error

	module := new(KafkaModule)
	module.brokers = brokers
	module.groupid = groupid
	module.logger = logger

	module.kExitChan = make(chan struct{})
	module.kPullChan = make(chan DoyoKafkaData, KREADCHAN_SIZE)
	module.kPushChan = make(chan DoyoKafkaData, KWRITECHAN_SIZE)

	if len(brokers) == 0 || -1 == strings.Index(brokers, ":") {
		err = fmt.Errorf("Invaid Parameter: %s is invalid", brokers)
		logger.Errorw("Invaid Parameter", "brokers", brokers)
		return nil, err
	}

	module.producer, err = kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokers})
	if err != nil {
		logger.Errorw("New kafka producer failed!", "error", err.Error())
		return nil, err
	} else {
		logger.Infof("New kafka producer %v", module.producer)

		module.exitWait.Add(1)
		go module.pushRoutine()
	}

	if len(topics) > 0 {
		// 创建consumer
		// earliest对于一个新的consumer，新group，会从最老的数据开始读取
		// https://kafka.apache.org/documentation/
		// session.timeout.ms: The minimum allowed session timeout for registered consumers.
		// Shorter timeouts result in quicker failure detection at the cost of more frequent consumer heartbeating,
		// which can overwhelm broker resources.
		module.consumer, err = kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers":  brokers,
			"group.id":           module.groupid,
			"socket.timeout.ms":  1000,
			"session.timeout.ms": 6000,
			"default.topic.config": kafka.ConfigMap{
				"auto.offset.reset": "earliest"}})

		if err != nil {
			logger.Errorw("New kafka consumer failed!", "error", err.Error())
			return nil, err
		} else {
			logger.Infof("New kafka consumer %v", module.consumer)
		}

		// 订阅
		err = module.consumer.SubscribeTopics(topics, func(my_c *kafka.Consumer, ev kafka.Event) error {
			// 这里仅仅log输出，confluentinc会自己做assign
			logger.Infof("SubscribeTopics kafka rebalance callback consumer[%s] ev[%s]", my_c.String(), ev.String())
			return nil
		})
		if err != nil {
			logger.Errorf("SubscribeTopics:%v failed! reason:%s", topics, err.Error())
			return nil, err
		}

		logger.Infof("SubscribeTopics:%v group.id[%s]", topics, module.groupid)

		module.exitWait.Add(1)
		go module.pullRoutine()
	}

	return module, nil
}

func (kfk *KafkaModule) PullChan() <-chan DoyoKafkaData {
	return kfk.kPullChan
}

func (kfk *KafkaModule) PushKfkData(topic string, writeData []byte) {
	kfk.kPushChan <- &DoyoKafkaWriteData{
		Topic:     topic,
		WriteData: writeData,
	}
}

func (kfk *KafkaModule) StopPull() {
	if kfk.consumer != nil {
		close(kfk.kExitChan)
	}
}

func (kfk *KafkaModule) ShutDown() {
	kfk.kPushChan <- &DoyoKafkaEofData{}
	kfk.exitWait.Wait()
}
