/*
 * @Author: calmwu
 * @Date: 2018-09-30 10:36:22
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-10-03 15:11:16
 */

package routerstub

import (
	"doyo-server-go/doyo-routersvr-go/proto"
	"errors"
	"time"
)

const (
	TopicRouterSvr = "DoyoRouterSvr"
)

// 业务服务类型名，！！！！规则：所有服务必须用Doyo开头！！！！！！
const (
	DoyoSvrTypeName_Business               = "DoyoBusinessSvr"
	DoyoSvrTypeName_Operational            = "DoyoOperationalSvr"
	DoyoSvrTypeName_Relationship           = "DoyoRelationShipSvr"
	DoyoSvrTypeName_RecommendedFriend      = "DoyoRecommendedFriendSvr"      // 好友推荐。推荐服务按产品设计分类、流水线串行、可拆可加
	DoyoSvrTypeName_RecommendedConcern     = "DoyoRecommendedConcernSvr"     // 关注推荐
	DoyoSvrTypeName_RecommendedPersonality = "DoyoRecommendedPersonalitySvr" // 个性推荐。国家、语言
)

var (
	ErrRouterStubWatiRespError = errors.New("RouterStub Wait Response failed")
	ErrRouterStubStop          = errors.New("RouterStub Has stopped")
	ErrRouterStubCallTimeOut   = errors.New("RouterStub Call TimeOut")
)

type OnReceive func(*RouterStubModule, string, string, []byte)

type rsCallerFutureResult struct {
	reponserTopic string
	resPayLoad    []byte
	err           error
}

type RSCallerFuture struct {
	timeout    time.Duration              // 超时间隔
	resultChan chan *rsCallerFutureResult // 结果返回通道
	MsgID      string                     // 请求的消息ID
	routerMsg  []byte                     // 请求的序列化消息
}

func (rscf *RSCallerFuture) init() {
	rscf.resultChan = make(chan *rsCallerFutureResult)
}

func (rscf *RSCallerFuture) response(respMsg *proto.RouterSvrDispatchMsg, err error) {
	var futureResult rsCallerFutureResult

	if err == nil {
		futureResult.reponserTopic = respMsg.FromTopic
		futureResult.resPayLoad = respMsg.PayLoad
	} else {
		futureResult.err = err
	}

	rscf.resultChan <- &futureResult
}

func (rscf *RSCallerFuture) WaitResp() (string, []byte, error) {
	result, ok := <-rscf.resultChan
	if ok {
		return result.reponserTopic, result.resPayLoad, result.err
	}
	return "", nil, ErrRouterStubWatiRespError
}
