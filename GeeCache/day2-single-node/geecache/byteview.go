package geecache

// ByteView 表示缓存值
type ByteView struct {
	// b 会存储真实的缓存值， byte类型是为了能够支持任意数据类型，字符串，图片
	b []byte
}

// Len 在lru中设置了value必须实现Value接口，Len() int 方法
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice b是只读的，所以返回一个拷贝，防止缓存值被外部程序修改
func (v ByteView) ByteSlice() []byte {
	return cloneByte(v.b)
}

// String 转换为字符串
func (v ByteView) String() string {
	return string(v.b)
}

func cloneByte(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
