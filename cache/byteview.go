/*
ByteView 只读数据结构，用于封装 []byte 类型
tip:
1. 使用 cloneBytes 函数来避免返回原始字节切片
2. 使用 String 方法来返回字符串表示
*/
package gocachex

// ByteView 是一个只读的数据结构，用于表示缓存值
// 它封装了 []byte 类型，实现了 Value 接口
// 所有返回的数据均为原始数据的副本，确保安全性
type ByteView struct {
	b []byte // 存储真实的字节数据
}

// Len 返回字节切片的长度
// 实现了 lru 包中 Value 接口所需的 Len() 方法
// 用于计算此值在缓存中占用的内存大小
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice 返回字节切片的副本
// 通过复制原始数据，确保返回的切片不会被外部修改
// 从而保证 ByteView 的只读特性
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// cloneBytes 创建并返回一个字节切片的深拷贝
// 此私有函数用于内部复制字节数据，避免共享内存
// 防止外部代码修改内部存储的数据
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

// String 将字节切片转换为字符串并返回
// 此方法方便将 ByteView 用于需要字符串的场合
// 直接调用 Go 标准库的 string() 转换函数
func (v ByteView) String() string {
	return string(v.b)
}
