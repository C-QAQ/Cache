package gocache

import (
	"gocache/lru"
	"sync"
)

type cache struct {
	mu         sync.Mutex // cache的全局锁
	lru        *lru.Cache // 为lru.Cache添加并发特性
	cacheBytes int64
}

// 向缓存中添加数据
func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil { // 延迟初始化，提高程序性能，减少程序对内存的要求
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

// 通过关键字从缓存中获取数据
func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil { // lru此时并没被初始化(这之前并没被添加过任何数据)
		return
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}

	return
}
