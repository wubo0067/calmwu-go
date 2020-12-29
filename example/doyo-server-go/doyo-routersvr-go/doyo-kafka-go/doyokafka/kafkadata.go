/*
 * @Author: calmwu
 * @Date: 2018-09-19 10:24:33
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-09-19 11:26:32
 */

package doyokafka

import "github.com/confluentinc/confluent-kafka-go/kafka"

type DoyoKafkaData interface {
	IsEof() bool
	Data() []byte
}

//---------------------------------------------------------------------
type DoyoKafkaEofData struct {
}

func (dked *DoyoKafkaEofData) IsEof() bool {
	return true
}

func (dked *DoyoKafkaEofData) Data() []byte {
	return nil
}

//---------------------------------------------------------------------
type DoyoKafkaWriteData struct {
	Topic     string
	WriteData []byte
}

func (dkwd *DoyoKafkaWriteData) IsEof() bool {
	return false
}

func (dkwd *DoyoKafkaWriteData) Data() []byte {
	return dkwd.WriteData
}

//---------------------------------------------------------------------
type DoyoKafkaReadData struct {
	TopicParition kafka.TopicPartition
	ReadData      []byte
}

func (dkrd *DoyoKafkaReadData) IsEof() bool {
	return false
}

func (dkrd *DoyoKafkaReadData) Data() []byte {
	return dkrd.ReadData
}

func (dkrd *DoyoKafkaReadData) FromInfo() *kafka.TopicPartition {
	return &dkrd.TopicParition
}
