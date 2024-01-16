/*
 * @Author: calm.wu
 * @Date: 2019-08-03 15:10:35
 * @Last Modified by: calmwu
 * @Last Modified time: 2020-08-15 20:16:25
 */

// Package task 任务对象
package task

import (
	"context"
	"fmt"
	"sync"

	"github.com/pkg/errors"

	"github.com/wubo0067/calmwu-go/utils"
)

// StepResult 任务执行的结果
type StepResult struct {
	StepName string
	Result   interface{}
	Err      error
}

// Step 任务执行的步骤
type Step interface {
	// Step 名字
	Name() string
	// 执行
	Do(ctx context.Context, stepIndex int, taskObj Task) *StepResult
	// 回滚
	Cancel(ctx context.Context, stepIndex int, taskObj Task) error
}

// TaskResult 任务执行的结果
type TaskResult struct {
	Result []*StepResult
}

// TaskEvent 通知事件
type TaskEvent struct {
	Info string
}

// TaskObserver 观察对象
type TaskObserver interface {
	OnNotify(*TaskEvent)
}

// Task 任务对象，管理 Step
type Task interface {
	// Name 任务名
	Name() string
	// Run 执行任务
	Run() (*TaskResult, error)
	// Rollback 任务回滚
	Rollback()
	// 得到运行参数
	GetTaskArgs() interface{}
	//
	GetStepResult(stepIndex int) *StepResult
}

var _ Task = &concreteTask{}

// ConcreteTask 具体的任务对象
type concreteTask struct {
	name          string          // 任务名字
	observer      TaskObserver    // 观察对象
	ctx           context.Context // 控制对象
	stepLst       []Step          // 步骤列表
	cancelStepLst []Step          // 回滚步骤列表
	taskArg       interface{}     // 任务的参数
	taskResult    TaskResult      // 任务执行的结果
	mutex         sync.Mutex      // 锁
	nc            utils.NoCopy
}

// MakeTask 构造一个 Task 对象
func MakeTask(ctx context.Context, name string, observer TaskObserver, taskArg interface{}, steps ...Step) (Task, error) {
	if taskArg == nil || len(steps) == 0 {
		return nil, errors.New("input parameters is invalid")
	}

	taskObj := &concreteTask{
		name:     name,
		ctx:      ctx,
		observer: observer,
		taskArg:  taskArg,
		stepLst:  steps,
	}
	return taskObj, nil
}

// Name 任务名
func (ti *concreteTask) Name() string {
	return ti.name
}

// Run 运行任务
func (ti *concreteTask) Run() (result *TaskResult, taskErr error) {
	result = nil
	ti.notifyObserver(fmt.Sprintf("Task:%s start running", ti.name))

	defer func() {
		if err := recover(); err != nil {
			stackInfo := utils.CallStack(3)
			taskErr = errors.Errorf("Panic! err:%v stack:%s", err, stackInfo)
		}
	}()

	for i, step := range ti.stepLst {
		// 执行到自己，回滚也从自己开始
		ti.cancelStepLst = append(ti.cancelStepLst, step)
		stepResult := step.Do(ti.ctx, i, ti)
		if stepResult.Err != nil {
			ti.notifyObserver(fmt.Sprintf("Task:%s step:%d name:%s execution failed", ti.name, i, step.Name()))
			return nil, stepResult.Err
		}
		ti.notifyObserver(fmt.Sprintf("Task:%s step:%d name:%s execution successed", ti.name, i, step.Name()))

		ti.taskResult.Result = append(ti.taskResult.Result, stepResult)
		select {
		case <-ti.ctx.Done():
			ti.notifyObserver(fmt.Sprintf("Task:%s was canceled after step:%d name:%s", ti.name, i, step.Name()))
			return nil, fmt.Errorf("task:%s was canceled after step:%d name:%s", ti.name, i, step.Name())
		default:
		}
	}
	ti.notifyObserver(fmt.Sprintf("Task:%s execution completed", ti.name))
	return &ti.taskResult, nil
}

// Rollback 任务回滚
func (ti *concreteTask) Rollback() {
	ti.notifyObserver(fmt.Sprintf("Task:%s start rollback", ti.name))
	cancelStepLstLen := len(ti.cancelStepLst)
	if cancelStepLstLen < 1 {
		return
	}
	for i := cancelStepLstLen - 1; i >= 0; i-- {
		step := ti.cancelStepLst[i]
		ti.notifyObserver(fmt.Sprintf("Task:%s step:%s start rollback operation", ti.name, step.Name()))
		step.Cancel(ti.ctx, i, ti)
		ti.notifyObserver(fmt.Sprintf("Task:%s step:%s rollback operation completed", ti.name, step.Name()))
	}
}

// GetTaskArgs 得到运行参数
func (ti *concreteTask) GetTaskArgs() interface{} {
	return ti.taskArg
}

func (ti *concreteTask) notifyObserver(info string) {
	if ti.observer != nil {
		ti.observer.OnNotify(&TaskEvent{
			Info: info,
		})
	}
}

// GetStepResult 得到前面一步的结果
func (ti *concreteTask) GetStepResult(stepIndex int) *StepResult {
	ti.mutex.Lock()
	defer ti.mutex.Unlock()
	if stepIndex < 0 || stepIndex >= len(ti.taskResult.Result) {
		return nil
	}
	return ti.taskResult.Result[stepIndex]
}
