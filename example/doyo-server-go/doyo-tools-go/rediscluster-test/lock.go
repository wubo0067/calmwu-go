/*
 * @Author: calmwu
 * @Date: 2018-10-27 22:34:41
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-10-29 14:02:29
 */

package main

import (
	base "doyo-server-go/doyo-base-go"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap/zapcore"
)

type GlobalLockState int

const (
	GlobalLockStatePreempting GlobalLockState = iota
	GloblaLockStateHold
	GlobalLockStateExpired
	GlobalLockStateFailure
)

var (
	releaseScript    string = "local fp = redis.pcall('get', KEYS[1]) if not fp or fp ~= ARGV[1] then return end return redis.pcall('del', KEYS[1])"
	releaseScriptSha string
)

type GlobalLock struct {
	ownerName             string
	redisClient           *redis.ClusterClient
	key                   string
	value                 string
	ttl                   time.Duration
	holdDuration          time.Duration
	retryTimes            int
	retryIntervalDuration time.Duration
	stateChan             chan GlobalLockState
	isHold                bool
	failureErr            error
}

func NewGlobalLock(ownerName string, redisClient *redis.ClusterClient, key string, holdDuration time.Duration,
	retryTimes int, retryIntvalDuration time.Duration) (*GlobalLock, error) {

	var err error
	if len(releaseScriptSha) == 0 {
		releaseScriptSha, err = redisClient.ScriptLoad(releaseScript).Result()
		if err != nil {
			base.ZLog.Errorf("ScriptLoad failed! reason:%s", err.Error())
			return nil, err
		} else {
			base.ZLog.Infof("ScriptLoad releaseScriptSha[%s]", releaseScriptSha)
		}
	}

	uid, _ := uuid.NewV4()

	globalLock := new(GlobalLock)
	globalLock.ownerName = ownerName
	globalLock.redisClient = redisClient
	globalLock.key = key
	globalLock.value = uid.String()
	globalLock.holdDuration = holdDuration
	globalLock.ttl = holdDuration + 5*time.Second          // key不可能没有ttl，如果所有者失效就只有依靠ttl了
	globalLock.retryTimes = retryTimes                     // 抢占失败后尝试次数，-1表示一直尝试
	globalLock.retryIntervalDuration = retryIntvalDuration // 尝试失败后的等待时间
	globalLock.stateChan = make(chan GlobalLockState, 16)  // 这里使用buffered channel
	globalLock.isHold = false
	globalLock.failureErr = nil

	return globalLock, err
}

func (gl *GlobalLock) DoPreempt(waitingLock bool) (bool, error) {
	if gl.isHold {
		err := fmt.Errorf("owner[%s] key[%s] Already hold lock", gl.ownerName, gl.key)
		base.ZLog.Errorf(err.Error())
		return false, err
	}

	if len(releaseScriptSha) == 0 {
		err := errors.New("Must load release script")
		base.ZLog.Error(err.Error())
		return false, err
	}

	waitingChan := make(chan struct{})

	go func() {
		for {
			// 开始抢占
			holdOK, err := gl.redisClient.SetNX(gl.key, gl.value, gl.ttl).Result()
			if err != nil {
				gl.failureErr = fmt.Errorf("owner[%s] setNX key[%s] value[%s] failed! reason:%s",
					gl.ownerName, gl.key, gl.value, err.Error())
				base.ZLog.Errorf(gl.failureErr.Error())
				gl.stateChan <- GlobalLockStateFailure
				close(waitingChan)
				return
			}

			if holdOK {
				// 锁已经持有
				gl.isHold = true
				base.ZLog.Infof("owner[%s] setNX key[%s] value[%s] Preemption successed!",
					gl.ownerName, gl.key, gl.value)
				gl.stateChan <- GloblaLockStateHold
				close(waitingChan)
				break
			}

			if gl.retryTimes == -1 {
				// 抢占失败后等待
				base.ZLog.Infof("owner[%s] setNX key[%s] value[%s] Preemption retry wait",
					gl.ownerName, gl.key, gl.value)
				gl.stateChan <- GlobalLockStatePreempting
				time.Sleep(gl.retryIntervalDuration)
			} else {
				gl.retryTimes--
				if gl.retryTimes == 0 {
					gl.failureErr = fmt.Errorf("owner[%s] setNX key[%s] value[%s] Preemption retry end",
						gl.ownerName, gl.key, gl.value)
					base.ZLog.Warnf(gl.failureErr.Error())
					gl.stateChan <- GlobalLockStateFailure
					close(waitingChan)
					return
				}
				base.ZLog.Debugf("owner[%s] setNX key[%s] value[%s] Preemption retry remain count[%d]",
					gl.ownerName, gl.key, gl.value, gl.retryTimes)
				gl.stateChan <- GlobalLockStatePreempting
				time.Sleep(gl.retryIntervalDuration)
			}
		}

		time.AfterFunc(gl.holdDuration, func() {
			// 锁持有到期，先释放在通知
			gl.Release()
			gl.stateChan <- GlobalLockStateExpired
		})
	}()

	if waitingLock {
		<-waitingChan
	}

	return gl.isHold, gl.failureErr
}

func (gl *GlobalLock) Result() (bool, error) {
	return gl.isHold, gl.failureErr
}

func (gl *GlobalLock) StateChan() <-chan GlobalLockState {
	return gl.stateChan
}

func (gl *GlobalLock) Release() {
	if !gl.isHold {
		base.ZLog.Debugf("owner[%s] not holding lock", gl.ownerName)
		return
	}

	if len(releaseScriptSha) == 0 {
		base.ZLog.Error("Must load release script")
		return
	}

	res, err := gl.redisClient.EvalSha(releaseScriptSha, []string{gl.key}, gl.value).Result()
	if err != nil {
		base.ZLog.Errorf("owner[%s] Release EvalSha key[%s] failed! reason:%s", gl.ownerName, gl.key, err.Error())
		return
	}

	if intRes, ok := res.(int64); !ok {
		base.ZLog.Errorf("owner[%s] Release EvalSha key[%s] failed!", gl.ownerName, gl.key)
	} else if intRes == 1 {
		gl.isHold = false
		base.ZLog.Infof("owner[%s] Release  EvalSha key[%s] successed!", gl.ownerName, gl.key)
	}
	return
}

func main() {

	base.InitDefaultZapLog("./lock_test.log", zapcore.DebugLevel)

	// lock 不支持
	redisdb := redis.NewClusterClient(&redis.ClusterOptions{
		// Addrs: []string{"192.168.68.228:7000", "192.168.68.228:7001", "192.168.68.229:7002",
		// 	"192.168.68.229:7003", "192.168.68.230:7004", "192.168.68.230:7005"},
		Addrs:       []string{"192.168.68.228:7000"},
		DialTimeout: 3 * time.Second,
		OnConnect: func(conn *redis.Conn) error {
			// 会连接所有的master
			callstack := base.CallStack(1)
			base.ZLog.Debugf("---------conn:%s, callstack:%s", conn.String(), callstack)
			return nil
		},
		ReadOnly: true, // 在从库上开启readonly
	})

	globalLock, err := NewGlobalLock("recommend", redisdb, "machine", 20*time.Second, 3, 2*time.Second)
	if err != nil {
		base.ZLog.Error(err.Error())
		return
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

	// 异步抢占
	globalLock.DoPreempt(false)
L:
	for {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			break L
		case state := <-globalLock.StateChan():
			switch state {
			case GlobalLockStatePreempting:
				base.ZLog.Info("GlobalLockStatePreempting")
			case GlobalLockStateExpired:
				base.ZLog.Info("GlobalLockStateExpired")
				// 继续竞争
				globalLock.DoPreempt(false)
			case GloblaLockStateHold:
				base.ZLog.Info("GloblaLockStateHold")
			case GlobalLockStateFailure:
				_, err := globalLock.Result()
				base.ZLog.Errorf("GlobalLockStateFailure err:%s", err.Error())
				break L
			}
		}
	}

	globalLock.Release()
	return
}
