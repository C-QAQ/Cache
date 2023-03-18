package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 将一段字节数据映射成一个正整数
type Hash func(data []byte) uint32

// Map 包含所有的散列键
type Map struct {
	hash     Hash           // hash函数，crc32哈希
	replicas int            // 虚拟节点的个数
	keys     []int          // 哈希环，环中元素是已经排序的
	hashMap  map[int]string // 虚拟节点和真实节点的映射关系
}

// New 创建一个Map实例，允许替换自定义的Hash函数
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	// 默认使用crc32
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 添加0个或多个真实节点的名称
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Get 选择节点的方法
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	// 顺时针找到第一个匹配的虚拟节点的下标
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[idx%len(m.keys)]]
}
