package singleflight

import "sync"

// 正在进行中或已经结束的请求
type call struct {
	wg  sync.WaitGroup // 加锁避免重入
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex // 对m的保护
	m  map[string]*call
}

/*
Do 方法，接收 2 个参数，第一个参数是 key，第二个参数是一个函数 fn。
Do 的作用就是，针对相同的 key，无论 Do 被调用多少次，函数 fn 都只会被调用一次，
等待 fn 调用结束了，返回返回值或错误。
*/
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock() // 如果请求正在进行中，则等待
		c.wg.Wait()
		return c.val, c.err // 请求结束，返回结果
	}
	c := new(call)
	c.wg.Add(1)  // 发起请求前加锁
	g.m[key] = c // 添加到g.m，表明key已经有对应的请求在处理
	g.mu.Unlock()

	c.val, c.err = fn() // 调用fn，发起请求
	c.wg.Done()         // 请求结束

	g.mu.Lock()
	delete(g.m, key) // 更新g.m
	g.mu.Unlock()

	return c.val, c.err // 返回结果
}
