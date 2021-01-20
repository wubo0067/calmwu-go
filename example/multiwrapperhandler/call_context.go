/*
 * @Author: calmwu
 * @Date: 2020-12-31 20:31:32
 * @Last Modified by: calmwu
 * @Last Modified time: 2020-12-31 21:35:56
 */

package main

import (
	"fmt"
	"math"
)

// 这就是abort的最大值，不可能超过这个的
const abortIndex int8 = math.MaxInt8 / 2

// HandlerFunc defines the handler used by gin middleware as return value.
type HandlerFunc func(*CallContext)

// HandlersChain defines a HandlerFunc array.
type HandlersChain []HandlerFunc

type CallContext struct {
	param    interface{}
	handlers HandlersChain
	index    int8
}

func (c *CallContext) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

func (c *CallContext) UseHandlers(handlers ...HandlerFunc) {
	c.handlers = append(c.handlers, handlers...)
}

func (c *CallContext) Abort() {
	// 通过设置超过上限的下标，进行流程控制
	c.index = abortIndex
}

func HandlerApple(c *CallContext) {
	fmt.Printf("enter ---HandlerApple--- param:%v\n", c.param)
	c.Next()
	defer fmt.Printf("exit ---HandlerApple--- param:%v\n", c.param)
}

func HandlerXiaomi(c *CallContext) {
	fmt.Printf("enter ---HandlerXiaomi--- param:%v\n", c.param)
	defer fmt.Printf("exit ---HandlerXiaomi--- param:%v\n", c.param)
	c.Abort()
}

func HandlerNvidia(c *CallContext) {
	fmt.Printf("enter ---HandlerNvidia--- param:%v\n", c.param)
	c.Next()
	defer fmt.Printf("exit ---HandlerNvidia--- param:%v\n", c.param)
}

func HandlerSony(c *CallContext) {
	fmt.Printf("enter ---HandlerSony--- param:%v\n", c.param)
	defer fmt.Printf("exit ---HandlerSony--- param:%v\n", c.param)
}

func testCallContext() {
	c := &CallContext{
		param: 3,
		index: -1,
	}

	c.UseHandlers(HandlerApple, HandlerXiaomi, HandlerNvidia, HandlerSony)
	c.Next()
}
