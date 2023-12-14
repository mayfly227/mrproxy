package common

import (
	"sync"
	"time"
)

// CacheKV Cache 结构表示一个简单的 K-V 缓存
type CacheKV struct {
	data     map[string]interface{}
	mutex    sync.RWMutex
	expire   time.Duration
	cleanup  *time.Ticker
	stopChan chan bool
}

// Set 将键值对存储到缓存中
func (c *CacheKV) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[key] = value
}

// Get 从缓存中获取键对应的值
func (c *CacheKV) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	value, ok := c.data[key]
	return value, ok
}

// Delete 从缓存中删除指定键的值
func (c *CacheKV) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.data, key)
}

// startCleanup 定期清理过期的缓存项
func (c *CacheKV) startCleanup() {
	for {
		select {
		case <-c.cleanup.C:
			c.mutex.Lock()
			for key := range c.data {
				// 如果某个键对应的值在指定的时间内未被访问，则删除它
				// 这里仅演示过期处理，实际使用中可能需要更复杂的策略
				delete(c.data, key)
			}
			c.mutex.Unlock()
		case <-c.stopChan:
			c.cleanup.Stop()
			return
		}
	}
}

// StopCleanup 停止定期清理
func (c *CacheKV) StopCleanup() {
	c.stopChan <- true
	close(c.stopChan)
}

// NewKVChche 创建一个新的缓存实例
func NewKVCache(expire time.Duration) *CacheKV {
	cache := &CacheKV{
		data:     make(map[string]interface{}),
		expire:   expire,
		cleanup:  time.NewTicker(expire),
		stopChan: make(chan bool),
	}

	go cache.startCleanup()

	return cache
}
