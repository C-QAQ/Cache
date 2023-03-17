package lru

import "container/list"

// Cache 遵循LRU缓存淘汰策略的没有并发安全的数据结构
type Cache struct {
	maxBytes int64                    // 最大容量
	nbytes   int64                    // 当前容量
	ll       *list.List               // 双端队列用于实现lru
	cache    map[string]*list.Element // key-value字典
	// 移除单个节点时运行的回调函数，可以为nil
	OnEvicted func(key string, value Value)
}

// 单个节点的数据类型
type entry struct {
	key   string // 淘汰缓存时用于删除对应的映射关系
	value Value
}

// Value 使用字节长度获取数据大小的所有数据类型
type Value interface {
	Len() int
}

// New Cache的构造函数
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Add 添加一个缓存到Cache
func (c *Cache) Add(key string, value Value) {
	// 判断是否已经存在这个缓存
	if ele, ok := c.cache[key]; ok { // 存在
		c.ll.MoveToFront(ele)                                  // 移动到队尾
		kv := ele.Value.(*entry)                               // 类型转换取出此缓存
		c.nbytes += int64(value.Len()) - int64(kv.value.Len()) // 更新当前Cache大小
		kv.value = value                                       // 更新缓存内容
	} else { // 不存在
		ele := c.ll.PushFront(&entry{key, value})        // 添加到队尾
		c.cache[key] = ele                               // 新添此缓存的映射关系
		c.nbytes += int64(len(key)) + int64(value.Len()) // 更新当前Cache大小
	}
	// 如果当前容量超过最大可支持容量，移除最少访问缓存
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// Get 通过关键字获取数据
func (c *Cache) Get(key string) (value Value, ok bool) {
	// 判断缓存是否存在
	if ele, ok := c.cache[key]; ok { // 存在
		c.ll.MoveToFront(ele)    // 将缓存移动到队尾
		kv := ele.Value.(*entry) // 取出此缓存
		return kv.value, true
	}
	return
}

// RemoveOldest 淘汰旧缓存
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back() // 取出对头
	if ele != nil {    // 对头存在
		c.ll.Remove(ele)                                       // 在双端队列中移除此缓存
		kv := ele.Value.(*entry)                               // 类型转换，取出此缓存的key-value对
		delete(c.cache, kv.key)                                // 在映射关系中删除此缓存
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len()) // 更新当前Cache的大小
		// 判断是否存在删除缓存时的操作
		if c.OnEvicted != nil { // 存在
			c.OnEvicted(kv.key, kv.value) // 执行此操作
		}
	}
}

// Len 返回缓存的数量
func (c *Cache) Len() int {
	return c.ll.Len()
}
