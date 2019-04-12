/*
 * @Author: calmwu
 * @Date: 2018-11-29 10:14:06
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-29 10:52:00
 */

// 这里没有使用WaitN，没有使用Context去停止

package utils

import (
	"time"

	"golang.org/x/time/rate"
)

type RateLimiter interface {
	// TryAccept return true if a token is taken immediately, it return false
	TryAccept() bool
	// Accept return once a token becomes avaliable
	Accept()
	// QPS return QPS of this rate limiter
	QPS() float32
}

type Clock interface {
	Now() time.Time
	Sleep(time.Duration)
}

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now()
}

func (realClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

type tokenBucketRateLimiter struct {
	limiter *rate.Limiter
	clock   Clock
	qps     float32
}

// qps 每秒的请求数量，也就是每秒的令牌生成数量
// burst 桶的深度
func NewTokenBucketRateLimiter(qps float32, burst int) RateLimiter {
	limiter := rate.NewLimiter(rate.Limit(qps), burst)
	return &tokenBucketRateLimiter{
		limiter: limiter,
		clock:   realClock{},
		qps:     qps,
	}
}

func NewTokenBucketRateLimiterWithClock(qps float32, burst int, c Clock) RateLimiter {
	limiter := rate.NewLimiter(rate.Limit(qps), burst)
	return &tokenBucketRateLimiter{
		limiter: limiter,
		clock:   c,
		qps:     qps,
	}
}

func (t *tokenBucketRateLimiter) TryAccept() bool {
	return t.TryAccept()
}

func (t *tokenBucketRateLimiter) Accept() {
	now := t.clock.Now()
	t.clock.Sleep(t.limiter.ReserveN(now, 1).DelayFrom(now))
}

func (t *tokenBucketRateLimiter) Stop() {}

func (t *tokenBucketRateLimiter) QPS() float32 {
	return t.qps
}
