// cache.go 文件实现了并发安全的缓存封装，通过互斥锁保护内部LRU缓存
// 对外暴露添加、获取缓存的功能，并确保在并发环境下的数据一致性
package gocachex

import (
	"goCacheX/lru"
	"sync"
)

// cache 是对LRU缓存的并发安全封装
// 内部使用互斥锁实现并发控制，保证在多线程环境下安全访问缓存
type cache struct {
	mu         sync.Mutex // 互斥锁，用于保证缓存操作的原子性
	lru        *lru.Cache // LRU缓存实例，存储实际的缓存数据
	cacheBytes int64      // 缓存的最大内存限制（字节）
}

// add 添加一个键值对到缓存
// 内部通过互斥锁保证并发安全，将操作委托给LRU缓存实现
// 参数:
//   - key: 缓存键
//   - value: 缓存值，为只读的ByteView类型
func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil { // 延迟初始化
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

// get 根据键获取缓存值
// 内部通过互斥锁保证并发安全，将查询委托给LRU缓存实现
// 返回:
//   - ByteView: 缓存的值，如果键不存在返回空ByteView
//   - bool: 表示键是否存在于缓存中
func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil { // 这个判断有必要，避免还没有初始化缓存时，调用get方法
		return
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), true
	}
	return
}

// Len 返回缓存中的元素数量
// 内部通过互斥锁保证并发安全，将查询委托给LRU缓存实现
// 返回:
//   - int: 缓存中的元素数量
func (c *cache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lru.Len()
}
