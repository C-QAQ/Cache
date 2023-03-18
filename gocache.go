package gocache

import (
	"fmt"
	"log"
	"sync"
)

// Group 一个缓存的命名空间
type Group struct {
	name      string
	getter    Getter // 缓存未命中的回调函数
	mainCache cache  // 支持并发
	peers     PeerPicker
}

// Getter 通过关键字，返回缓存的字节类型
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 使用函数实现Getter
type GetterFunc func(key string) ([]byte, error)

// Get 实现Getter接口函数
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex              // 全局锁
	groups = make(map[string]*Group) // 关键字和对应缓存的映射
)

// NewGroup 创建一个Group实例
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil { // 未指定缓存未命中时的回调函数
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g // 添加映射关系
	return g
}

// GetGroup 通过name返回对应Group，若没有则返回nil
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get 通过key获取缓存的字节类型
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" { // 不存在key为空的缓存
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok { // 缓存击中，且maincache实现了并发获取，避免了读写冲突
		log.Println("[GoCache] hit")
		return v, nil
	}

	return g.load(key) // 缓存不存在
}

// RegisterPeers 注册一个PeerPicker用于选择远端节点
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) load(key string) (value ByteView, err error) {
	if g.peers != nil {
		if peer, ok := g.peers.PickPeer(key); ok {
			if value, err = g.getFromPeer(peer, key); err == nil {
				return value, nil
			}
			log.Println("[GoCache] Failed to get from peer", err)
		}
	}
	return g.getLocally(key) // 远端获取失败，从本地获取对应数据
}

// 使用实现了PeerGetter接口的httpGetter从访问远端节点获取缓存值
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key) // 执行用户自定的缓存未命中的回调函数
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value) // 此缓存未命中但是被用户查询，所以需要再次添加到缓存中
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
