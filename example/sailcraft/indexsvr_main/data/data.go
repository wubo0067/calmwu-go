/*
 * @Author: calmwu
 * @Date: 2017-09-18 14:27:09
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-09-19 15:00:51
 * @Comment:
 */

package data

import (
	"sailcraft/base"
	"sailcraft/indexsvr_main/proto"

	"github.com/emirpasic/gods/lists/singlylinkedlist"
)

type DataActionType int

const (
	E_DATAACTION_LIKE = iota
	E_DATAACTION_MATCH
	E_DATAACTION_SET
	E_DATAACTION_TRUNCATE
	E_DATAACTION_DEL
)

type DataMgrI interface {
	// 数据加载
	Load() error
	// 添加
	Set(key string, data proto.DataMetaI)
	// key匹配数据查询
	Like(key string, dataSetType proto.DataSetType, queryCount int) *singlylinkedlist.List
	// key完全匹配查询
	Match(key string, dataSetType proto.DataSetType, queryCount int) *singlylinkedlist.List
	// 删除key
	Delete(key string, data proto.DataMetaI)
	// 修改key
	Modify(oldKey, newKey string, oldData, newData proto.DataMetaI)
	// 清除所有数据
	Truncate()
	// 重新加载数据
	Reload() error
}

var (
	GDataMgr DataMgrI = nil
)

// 数据查询结果
type QueryResultS struct {
	Ok     bool
	Result *singlylinkedlist.List
}

// 自定义数据操作
type CustomDataAction func(client, userData interface{}) (*QueryResultS, error)

// 数据操作
type DataActionInfoS struct {
	actionType  DataActionType
	key         string
	value       proto.DataMetaI
	resultChan  chan<- *QueryResultS
	dataSetType proto.DataSetType
}

func CreateDataMgr(name string) DataMgrI {
	if GDataMgr == nil {
		switch name {
		case "redis":
			GDataMgr = new(RedisDataMgr)
		default:
			base.GLog.Error("Unknown name[%s]", name)
		}
	}

	return GDataMgr
}
