package common

import (
	"container/list"
	"errors"
	"sync"
)

// CacheNode LRU Node
type CacheNode struct {
	Key, Value interface{}
}

// NewCacheNode 创建一个节点指针
func (cnode *CacheNode) NewCacheNode(k, v interface{}) *CacheNode {
	return &CacheNode{k, v}
}

// LRUCache 整体封装结构
type LRUCache struct {
	Capacity int
	dlist    *list.List
	cacheMap map[interface{}]*list.Element
	cacheMutex sync.Mutex
}

// NewLRUCache 创建一个整体封装结构指针
func NewLRUCache(cap int) *LRUCache {
	return &LRUCache{
		Capacity: cap,
		dlist:    list.New(),
		cacheMap: make(map[interface{}]*list.Element)}
}

// Size 返回目前的容量
func (lru *LRUCache) Size() int {
	return lru.dlist.Len()
}

// Set 插入一个节点
func (lru *LRUCache) Set(k, v interface{}) error {
	lru.cacheMutex.Lock()
	defer lru.cacheMutex.Unlock()

	if lru.dlist == nil {
		return errors.New("LRUCache Not Initial")
	}

	if pElement, ok := lru.cacheMap[k]; ok {
		lru.dlist.MoveToFront(pElement)
		pElement.Value.(*CacheNode).Value = v
		return nil
	}

	newElement := lru.dlist.PushFront(&CacheNode{k, v})
	lru.cacheMap[k] = newElement

	if lru.dlist.Len() > lru.Capacity {
		//移掉最后一个
		lastElement := lru.dlist.Back()
		if lastElement == nil {
			return nil
		}
		cacheNode := lastElement.Value.(*CacheNode)
		delete(lru.cacheMap, cacheNode.Key)
		lru.dlist.Remove(lastElement)
	}
	return nil
}

// Get 返回节点
func (lru *LRUCache) Get(k interface{}) (v interface{}, ret bool, err error) {
	lru.cacheMutex.Lock()
	defer lru.cacheMutex.Unlock()

	if lru.cacheMap == nil {
		return v, false, errors.New("LRUCache Not Initial")
	}

	if pElement, ok := lru.cacheMap[k]; ok {
		lru.dlist.MoveToFront(pElement)
		return pElement.Value.(*CacheNode).Value, true, nil
	}
	return v, false, nil
}

// Remove 删除节点
func (lru *LRUCache) Remove(k interface{}) bool {
	lru.cacheMutex.Lock()
	defer lru.cacheMutex.Unlock()
	
	if lru.cacheMap == nil {
		return false
	}

	if pElement, ok := lru.cacheMap[k]; ok {
		cacheNode := pElement.Value.(*CacheNode)
		delete(lru.cacheMap, cacheNode.Key)
		lru.dlist.Remove(pElement)
		return true
	}
	return false
}
