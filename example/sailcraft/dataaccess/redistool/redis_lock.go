package redistool

import (
	"fmt"
	"sailcraft/base"
	"time"

	"github.com/satori/go.uuid"
)

var (
	SafeUnlockRoutine     string = "local fp = redis.pcall('get', KEYS[1]) if not fp or fp ~= ARGV[1] then return end return redis.pcall('del', KEYS[1])"
	GSafeUnlockRoutineSHA []byte = nil
)

const (
	//这个锁采用自旋锁实现，高效.
	//为了避免异常时死锁，这里设置下最大的自旋次数
	//kSnoozeTime, 这里设置100毫秒旋一次
	SnoozeTime   = 100
	TryLockTimes = 20
)

func getSafeUnlockRoutineSHA() ([]byte, error) {
	if GSafeUnlockRoutineSHA != nil {
		return GSafeUnlockRoutineSHA, nil
	}

	redisNode, err := GRedisManager.GetSingletonRedisMgr()
	if redisNode == nil {
		return nil, err
	}

	fp, err := redisNode.ScriptLoad([]byte(SafeUnlockRoutine))
	if err == nil {
		GSafeUnlockRoutineSHA = fp.([]byte)
		return GSafeUnlockRoutineSHA, nil
	}

	return nil, err
}

func SafeLock(key string, value string, ttl int) error {
	base.GLog.Debug("SafeLock enter key %s value %s ttl %d", key, value, ttl)

	// 必须传ttl过来, 默认填0,
	if ttl == 0 {
		ttl = 5
	}

	redisNode, err := GRedisManager.GetSingletonRedisMgr()
	if redisNode == nil {
		return err
	}

	err = redisNode.StringSetNX(key, []byte(value), ttl)
	if err != nil {
		return err
	}

	return nil
}

func loopLock(ch chan int, key string, value string, ttl int) {
	i := 1
	for i <= TryLockTimes {
		i = i + 1

		err := SafeLock(key, value, ttl)
		if err == nil {
			ch <- 0
			return
		}

		time.Sleep(SnoozeTime * time.Millisecond)
	}

	return
}

func SpinLock(key string, value string, ttl int) error {
	base.GLog.Debug("SpinLock enter key %s value %s ttl %d", key, value, ttl)

	result := make(chan int)
	go loopLock(result, key, value, ttl)

	select {
	case _, ok := <-result:
		if ok {
			return nil
		}
	case <-time.After(2 * time.Second):
		return fmt.Errorf("SpinLock failed")
	}

	return fmt.Errorf("SpinLock failed")
}

func SpinLockWithFingerPoint(key string, ttl int) (string, error) {
	u2, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("gen uuid failed")
	}

	value := u2.String()

	err = SpinLock(key, value, ttl)
	if err != nil {
		return "", err
	}

	return value, nil
}

func UnLock(key string, value string) error {
	base.GLog.Debug("UnLock enter key %s value %s", key, value)

	redisNode, err := GRedisManager.GetSingletonRedisMgr()
	if redisNode == nil {
		return err
	}

	sha, err := getSafeUnlockRoutineSHA()
	if sha == nil {
		return err
	}

	args := make([]interface{}, 0)
	args = append(args, sha)
	args = append(args, "1")
	args = append(args, key)
	args = append(args, value)

	_, err = redisNode.Evalsha(args)
	if err != nil {
		return err
	}

	return nil
}
