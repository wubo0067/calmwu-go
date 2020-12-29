/*
 * @Author: calmwu
 * @Date: 2018-09-27 20:14:31
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-10-03 15:10:14
 */

package routerstub

import (
	base "doyo-server-go/doyo-base-go"
	"doyo-server-go/doyo-routersvr-go/proto"
	"time"

	"github.com/pquerna/ffjson/ffjson"

	uuid "github.com/satori/go.uuid"
)

const (
	callFutureWaitDuration = 3 * time.Second
)

// rsm: 调用对象
// fromTopic: 发送方的kafka topic
// toSvrType: 消息投递的目标服务类型
// dispatchPolicy: 消息转发策略
// dispatchPolicyArg: 转发使用的参数，例如uin。dest-hash(arg) mod servInstCount
// payLoad: 序列化后的数据
// 返回值：msgID, error
func Notify(rsm *RouterStubModule, fromTopic string, toSvrType string, dispatchPolicy proto.RouterSvrDispatchPolicy,
	dispatchPolicyArg string, payLoad []byte) (msgID string, err error) {

	uid, _ := uuid.NewV4()
	msgID = uid.String()

	msg := &proto.RouterSvrDispatchMsg{
		MessageID:         msgID,
		FromTopic:         fromTopic,
		ToServType:        toSvrType,
		DispatchPolicy:    dispatchPolicy,
		DispatchPolicyArg: dispatchPolicyArg,
		PayLoad:           payLoad,
	}

	// 序列化
	jsonData, err := ffjson.Marshal(msg)
	if err != nil {
		base.ZLog.Error("ffjson Marshal failed! reason:%s", err.Error())
		return msgID, err
	}

	// 发送
	rsm.Notify(TopicRouterSvr, jsonData)

	return msgID, nil
}

// rpc call
func Call(rsm *RouterStubModule, callerTopic string, toSvrType string,
	dispatchPolicy proto.RouterSvrDispatchPolicy, dispatchPolicyArg string,
	reqPayLoad []byte, timeout time.Duration) (string, string, []byte, error) {

	uid, _ := uuid.NewV4()
	msgID := uid.String()

	msg := &proto.RouterSvrDispatchMsg{
		MessageID:         msgID,
		FromTopic:         callerTopic,
		ToServType:        toSvrType,
		DispatchPolicy:    dispatchPolicy,
		DispatchPolicyArg: dispatchPolicyArg,
		PayLoad:           reqPayLoad,
	}

	// 序列化
	jsonData, err := ffjson.Marshal(msg)
	if err != nil {
		base.ZLog.Error("ffjson Marshal failed! reason:%s", err.Error())
		return msgID, "", nil, err
	}

	if timeout < callFutureWaitDuration {
		timeout = callFutureWaitDuration
	}
	callFuture, _ := rsm.Call(msgID, jsonData, timeout)
	base.ZLog.Debugf("Call MsgID[%s] callerTopic[%s] toSvrType[%s]", msgID, callerTopic, toSvrType)

	responserTopic, resPayLoad, err := callFuture.WaitResp()
	return callFuture.MsgID, responserTopic, resPayLoad, err
}

// call-reply
func Reply(rsm *RouterStubModule, destSvrTopic string, msgID string, fromTopic string, payLoad []byte) error {
	msg := &proto.RouterSvrDispatchMsg{
		MessageID: msgID,
		FromTopic: fromTopic,
		PayLoad:   payLoad,
	}

	// 序列化
	jsonData, err := ffjson.Marshal(msg)
	if err != nil {
		base.ZLog.Error("ffjson Marshal failed! reason:%s", err.Error())
		return err
	}

	// 发送
	rsm.Notify(destSvrTopic, jsonData)
	base.ZLog.Debugf("Reply MsgID[%s] destSvrTopic[%s] fromTopic[%s]", msgID, destSvrTopic, fromTopic)

	return nil
}
