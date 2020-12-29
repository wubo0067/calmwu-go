/*
 * @Author: CALM.WU
 * @Date: 2020-05-31 15:45:47
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2020-05-31 16:14:29
 */

package main

import (
	"context"
	"fmt"
)

// Client 调用对象
type Client interface {
	Call(ctx context.Context, req interface{}, res interface{}) error
}

// Wrapper是组装函数
type Wrapper func(c Client) Client

//--------------------------------------------
type DefaultClientImpl struct{}

func (c *DefaultClientImpl) Call(ctx context.Context, req interface{}, res interface{}) error {
	fmt.Println("Default Client implement")
	return nil
}

//--------------------------------------------
// 这些wrapper其实就是实现了Client接口
type clientSelectWrapper struct {
	Client
}

func (c *clientSelectWrapper) Call(ctx context.Context, req interface{}, res interface{}) error {
	fmt.Println("client Select Wrapper")
	if c.Client != nil {
		return c.Client.Call(ctx, req, res)
	}
	return nil
}

// NewSelectClientWrapper 辅助函数
func NewSelectClientWrapper() Wrapper {
	return func(c Client) Client {
		return &clientSelectWrapper{c}
	}
}

//--------------------------------------------
type clientTraceWrapper struct {
	Client
}

func (c *clientTraceWrapper) Call(ctx context.Context, req interface{}, res interface{}) error {
	fmt.Println("client Trace Wrapper")
	if c.Client != nil {
		return c.Client.Call(ctx, req, res)
	}
	return nil
}

// NewTraceClientWrapper 辅助函数
func NewTraceClientWrapper() Wrapper {
	return func(c Client) Client {
		return &clientTraceWrapper{c}
	}
}

//--------------------------------------------

func testWrapClientCall(w ...Wrapper) {
	var client Client = &DefaultClientImpl{}

	for i := len(w); i > 0; i-- {
		// 组合起来
		client = w[i-1](client)
	}

	client.Call(context.TODO(), nil, nil)
}

func testWrapCall() {
	testWrapClientCall(NewSelectClientWrapper(), NewTraceClientWrapper())
}
