/*
 * @Author: calmwu
 * @Date: 2018-11-29 10:44:51
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-29 11:25:32
 */

package base

import (
	"testing"
)

func TestThrottle(t *testing.T) {
	logger := NewSimpleLog(nil)

	rateLimiter := NewTokenBucketRateLimiter(float32(2), 5)

	logger.Printf("QPS:%f\n", rateLimiter.QPS())

	acceptCount := 0
	for acceptCount < 10 {
		rateLimiter.Accept()
		logger.Printf("%d can do-------------\n", acceptCount)

		acceptCount++
	}
}

func TestMultiThreadThrottle(t *testing.T) {
	logger := NewSimpleLog(nil)

	rateLimiter := NewTokenBucketRateLimiter(float32(2), 5)

	logger.Printf("QPS:%f\n", rateLimiter.QPS())

	parallelSize := 20
	notify := make(chan int, parallelSize)
	index := 0
	for index < parallelSize {
		go func(i int) {
			rateLimiter.Accept()
			logger.Printf("index:%d can do--------\n", i)

			notify <- i
		}(index)
		index++
	}

	totalNum := 0
	receiveCount := 0
	for i := range notify {
		totalNum += i
		receiveCount++
		if receiveCount == 20 {
			logger.Printf("receive:%d completed, totalNum:%d", receiveCount, totalNum)
			break
		}
	}
}
