/*
 * @Author: calm.wu
 * @Date: 2019-08-03 17:13:43
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-08-03 19:46:26
 */

package main

import (
	"context"
	"log"

	"github.com/wubo0067/calmwu-go/utils/task"
)

type myTaskObserver struct{}

func (mto myTaskObserver) OnNotify(event *task.TaskEvent) {
	log.Println("observer:", event.Info)
}

type AddStep struct {
	increaseNum int
}

func (as *AddStep) Name() string {
	return "AddStep"
}

func (as *AddStep) Do(ctx context.Context, stepIndex int, taskObj task.Task) *task.StepResult {
	log.Printf("StepIndex:%d AddStep Do\n", stepIndex)
	args := taskObj.GetTaskArgs().(int)
	args = args + as.increaseNum
	log.Printf("AddStep Do result:%d\n", args)
	return &task.StepResult{
		StepName: "AddStep",
		Result:   args,
		Err:      nil,
	}
}

func (as *AddStep) Cancel(ctx context.Context, stepIndex int, taskObj task.Task) error {
	log.Println("AddStep Cancel")
	res := taskObj.GetStepResult(stepIndex).Result.(int) - as.increaseNum
	log.Printf("AddStep Cancel result:%d\n", res)
	return nil
}

type MultiStep struct {
	multiNum int
}

func (ms *MultiStep) Name() string {
	return "MultiStep"
}

func (ms *MultiStep) Do(ctx context.Context, stepIndex int, taskObj task.Task) *task.StepResult {
	log.Printf("StepIndex:%d MultiStep Do\n", stepIndex)
	args := taskObj.GetStepResult(stepIndex - 1).Result.(int)
	args = args * ms.multiNum
	log.Printf("MultiStep Do result:%d\n", args)
	return &task.StepResult{
		StepName: "MultiStep",
		Result:   args,
		Err:      nil,
	}
}

func (ms *MultiStep) Cancel(ctx context.Context, stepIndex int, taskObj task.Task) error {
	log.Println("MultiStep Cancel")
	res := taskObj.GetStepResult(stepIndex).Result.(int) / ms.multiNum
	log.Printf("MultiStep Cancel result:%d\n", res)
	return nil
}

func main() {
	ctx, _ := context.WithCancel(context.Background())
	taskObj, err := task.MakeTask(ctx, "calcTask", &myTaskObserver{}, 90, &AddStep{10}, &MultiStep{6})
	if err != nil {
		return
	}

	result, err := taskObj.Run()
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("result:%+v\n", result)

	taskObj.Rollback()

	labels := map[string]string{
		"hello": "world",
	}
}
