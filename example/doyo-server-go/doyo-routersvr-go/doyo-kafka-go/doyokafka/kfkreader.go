/*
 * @Author: calmwu
 * @Date: 2018-09-15 16:31:33
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-09-25 16:06:39
 */

package doyokafka

import (
	base "doyo-server-go/doyo-base-go"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func (km *KafkaModule) pullRoutine() {
	km.logger.Infof("kafkaModule pullRoutine running")

	defer func() {
		km.exitWait.Done()
		if err := recover(); err != nil {
			// 回收panic，防止整个服务panic
			stackInfo := base.CallStack(1)
			km.logger.DPanicw("pullRoutine panic recovered! err", err, "stack", string(stackInfo))
		}
	}()
L:
	for {
		select {
		case <-km.kExitChan:
			km.logger.Info("pullRoutine receive exit noitfy")
			// 发送结束包
			km.kPullChan <- &DoyoKafkaEofData{}
			// 取消订阅
			km.consumer.Unsubscribe()
			break L
		default:
			ev := km.consumer.Poll(500)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				km.logger.Debugw("Receive msg from kafka", "topicPartition", e.TopicPartition.String())
				// 消息写入业务可读管道
				km.kPullChan <- &DoyoKafkaReadData{TopicParition: e.TopicPartition, ReadData: e.Value}
			case kafka.PartitionEOF:
				km.logger.Debugf("Reached %s", e.String())
			case kafka.Error:
				km.logger.Errorf("Error %s", e.String())
			case kafka.OffsetsCommitted:
				km.logger.Debugf("Commit %s", e.String())
			case kafka.AssignedPartitions:
				km.logger.Debugf("AssignedPartitions %s", e.String())
			default:
				km.logger.Debugf("Unhandled event %T %v", e, e)
			}
		}
	}

	km.logger.Info("kafkaModule pullRoutine exit!")
}
