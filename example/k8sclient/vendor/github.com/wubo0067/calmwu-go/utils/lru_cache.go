/*
 * @Author: calm.wu
 * @Date: 2019-07-15 09:53:05
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-07-15 11:04:11
 */

package utils

import (
	"errors"
	"fmt"
	"sync"
)

// 节点
type LRUCacheNode struct {
	Key   string
	Value interface{}

	prevNode *LRUCacheNode
	nextNode *LRUCacheNode
}

type LRUCacheFunctions interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{})
	Clear()
}

type LRUCache struct {
	Capacity int
	// 存储要用二维关系，map + 链表
	CacheStorage map[string]*LRUCacheNode

	firstNode *LRUCacheNode
	lastNode  *LRUCacheNode

	//
	monitor sync.Mutex

	// 匿名接口
	LRUCacheFunctions
}

// 构建一个LRUCache对象
func NewLRUCache(capacity int) (*LRUCache, error) {
	if capacity <= 0 {
		return nil, errors.New("capactiy must be larger than zero")
	} else {
		cache := &LRUCache{
			Capacity:     capacity,
			CacheStorage: make(map[string]*LRUCacheNode),
			firstNode:    nil,
			lastNode:     nil,
		}
		return cache, nil
	}
}

func (lc *LRUCache) pushToHead(key string, value interface{}) {
	node := &LRUCacheNode{
		Key:      key,
		Value:    value,
		nextNode: lc.firstNode,
		prevNode: nil,
	}

	if lc.firstNode != nil {
		lc.firstNode.prevNode = node
	}

	if lc.lastNode == nil {
		lc.lastNode = node
	}

	lc.firstNode = node
	lc.CacheStorage[key] = node
}

func (lc *LRUCache) removeNode(key string) {
	if node, exists := lc.CacheStorage[key]; !exists {
		// 如果不存在
		return
	} else {
		if node.prevNode != nil {
			node.prevNode.nextNode = node.nextNode
		} else {
			lc.firstNode = node.nextNode
		}

		if node.nextNode != nil {
			node.nextNode.prevNode = node.prevNode
		} else {
			lc.lastNode = node.prevNode
		}

		delete(lc.CacheStorage, key)
	}
}

// 通过key获取value，同时调整node位置
func (lc *LRUCache) Get(key string) (interface{}, error) {
	lc.monitor.Lock()
	defer lc.monitor.Unlock()

	if cacheNode, exists := lc.CacheStorage[key]; !exists {
		return nil, fmt.Errorf("key[%s] is invalid", key)
	} else {
		lc.removeNode(key)
		lc.pushToHead(key, cacheNode.Value)
		return cacheNode.Value, nil
	}
}

func (lc *LRUCache) Set(key string, value interface{}) {
	lc.monitor.Lock()
	defer lc.monitor.Unlock()

	// 删除后插入
	lc.removeNode(key)
	lc.pushToHead(key, value)

	if len(lc.CacheStorage) > lc.Capacity {
		// 直接删除末尾的
		lc.removeNode(lc.lastNode.Key)
	}
}

func (lc *LRUCache) Clear() {
	lc.monitor.Lock()
	defer lc.monitor.Unlock()

	for key, _ := range lc.CacheStorage {
		lc.removeNode(key)
	}
	lc.firstNode = nil
	lc.lastNode = nil
}
