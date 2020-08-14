/*
 * @Author: calmwu
 * @Date: 2018-10-29 14:06:53
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-29 10:44:29
 */

package redistool

import (
	"context"
	"errors"
	"fmt"
	"time"

	utils "github.com/wubo0067/calmwu-go/utils"

	"github.com/go-redis/redis"
	uuid "github.com/satori/go.uuid"
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
	ownerName string
	//redisClient           *redis.ClusterClient
	redisCmd              redis.Cmdable
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

func NewGlobalLock(ownerName string, redisCmd redis.Cmdable, key string, holdDuration time.Duration,
	retryTimes int, retryIntvalDuration time.Duration) (*GlobalLock, error) {

	var err error
	if len(releaseScriptSha) == 0 {
		releaseScriptSha, err = redisCmd.ScriptLoad(context.TODO(), releaseScript).Result()
		if err != nil {
			utils.ZLog.Errorf("ScriptLoad failed! reason:%s", err.Error())
			return nil, err
		} else {
			utils.ZLog.Infof("ScriptLoad releaseScriptSha[%s]", releaseScriptSha)
		}
	}

	uid, _ := uuid.NewV4()

	globalLock := new(GlobalLock)
	globalLock.ownerName = ownerName
	globalLock.redisCmd = redisCmd
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
		utils.ZLog.Errorf(err.Error())
		return false, err
	}

	if len(releaseScriptSha) == 0 {
		err := errors.New("Must load release script")
		utils.ZLog.Errorf(err.Error())
		return false, err
	}

	waitingChan := make(chan struct{})

	go func() {
		for {
			// 开始抢占
			holdOK, err := gl.redisCmd.SetNX(context.TODO(), gl.key, gl.value, gl.ttl).Result()
			if err != nil {
				gl.failureErr = fmt.Errorf("owner[%s] setNX key[%s] value[%s] failed! reason:%s",
					gl.ownerName, gl.key, gl.value, err.Error())
				utils.ZLog.Errorf(gl.failureErr.Error())
				gl.stateChan <- GlobalLockStateFailure
				close(waitingChan)
				return
			}

			if holdOK {
				// 锁已经持有
				gl.isHold = true
				utils.ZLog.Infof("owner[%s] setNX key[%s] value[%s] Preemption successed!",
					gl.ownerName, gl.key, gl.value)
				gl.stateChan <- GloblaLockStateHold
				close(waitingChan)
				break
			}

			if gl.retryTimes == -1 {
				// 抢占失败后等待
				// utils.ZLog.Infof("owner[%s] setNX key[%s] value[%s] Preemption retry wait",
				// 	gl.ownerName, gl.key, gl.value)
				gl.stateChan <- GlobalLockStatePreempting
				time.Sleep(gl.retryIntervalDuration)
			} else {
				gl.retryTimes--
				if gl.retryTimes == 0 {
					gl.failureErr = fmt.Errorf("owner[%s] setNX key[%s] value[%s] Preemption retry end",
						gl.ownerName, gl.key, gl.value)
					utils.ZLog.Warnf(gl.failureErr.Error())
					gl.stateChan <- GlobalLockStateFailure
					close(waitingChan)
					return
				}
				// utils.ZLog.Debugf("owner[%s] setNX key[%s] value[%s] Preemption retry remain count[%d]",
				// 	gl.ownerName, gl.key, gl.value, gl.retryTimes)
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
		utils.ZLog.Debugf("owner[%s] not holding lock", gl.ownerName)
		return
	}

	if len(releaseScriptSha) == 0 {
		utils.ZLog.Errorf("Must load release script")
		return
	}

	res, err := gl.redisCmd.EvalSha(context.TODO(), releaseScriptSha, []string{gl.key}, gl.value).Result()
	if err != nil {
		utils.ZLog.Errorf("owner[%s] Release EvalSha key[%s] failed! reason:%s", gl.ownerName, gl.key, err.Error())
		return
	}

	if intRes, ok := res.(int64); !ok {
		utils.ZLog.Errorf("owner[%s] Release EvalSha key[%s] failed!", gl.ownerName, gl.key)
	} else if intRes == 1 {
		gl.isHold = false
		utils.ZLog.Infof("owner[%s] Release  EvalSha key[%s] successed!", gl.ownerName, gl.key)
	}
	return
}
