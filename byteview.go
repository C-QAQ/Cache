package gocache

// ByteView 只读的字节类型视图
type ByteView struct {
	b []byte // 存储真实的缓存值，能够支持任意数据类型的存储(字符串、图片等)
}

// Len 返回字节视图的大小
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice 以切片类型返回字节内容
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// String 以字符串类型返回字节视图的内容
func (v ByteView) String() string {
	return string(v.b)
}

// cloneBytes 返回一个字节内容的副本
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
