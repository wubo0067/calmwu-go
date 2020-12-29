/*
 * @Author: calmwu
 * @Date: 2018-09-15 18:13:11
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-09-27 19:29:38
 */

package doyokafka

import (
	base "doyo-server-go/doyo-base-go"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const (
	ProducerDeliverChanSize = 1024
)

func (km *KafkaModule) pushRoutine() {
	km.logger.Infof("kafkaModule pushRoutine running")

	defer func() {
		km.exitWait.Done()
		if err := recover(); err != nil {
			// 回收panic，防止整个服务panic
			stackInfo := base.CallStack(1)
			km.logger.DPanicw("pushRoutine panic recovered! err", err, "stack", string(stackInfo))
		}
	}()

	// producer deliver event chan
	producerDeliverChan := make(chan kafka.Event, ProducerDeliverChanSize)

L:
	for {
		select {
		case ev, ok := <-producerDeliverChan:
			if ok {
				if msg, ok := ev.(*kafka.Message); ok {
					if msg.TopicPartition.Error != nil {
						// TODO: 这里要使用future来给业务层通报发送错误
						km.logger.Errorf("pushRoutine delivery failed: %v", msg.TopicPartition.Error)
					} else {
						km.logger.Debugf("pushRoutine delivery msg to topic %s partition %d at offset %v",
							*msg.TopicPartition.Topic, msg.TopicPartition.Partition, msg.TopicPartition.Offset)
					}
				}
			}
		case kdata, ok := <-km.kPushChan:
			if ok {
				// 发送出去
				switch kd := kdata.(type) {
				case *DoyoKafkaEofData:
					km.logger.Info("pushRoutine receive eof message")
					break L
				case *DoyoKafkaWriteData:
					kMsg := &kafka.Message{
						TopicPartition: kafka.TopicPartition{Topic: &kd.Topic, Partition: kafka.PartitionAny},
						Value:          kd.Data(),
						Headers:        []kafka.Header{kafka.Header{Key: "doyo-kafka-go", Value: []byte("doyo-kafka-go-kwriter")}},
					}
					err := km.producer.Produce(kMsg, producerDeliverChan)
					if err != nil {
						km.logger.Errorf("pushRoutine delivery failed: %s", err.Error())
					}
				}

			} else {
				km.logger.Error("pushRoutine read kPushChan failed!")
			}
		}
	}
	km.logger.Info("pushRoutine exit!")
}
