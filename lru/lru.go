// Package lru 实现了一个LRU（最近最少使用）缓存数据结构。
//
// LRU缓存具有固定的最大内存限制，当内存占用超过这个限制时，
// 会自动淘汰最久未使用的缓存项。这种策略在有限内存环境下
// 特别有效，既能保证热点数据的快速访问，又能有效管理内存资源。
//
// 实现原理：
// 1. 使用双向链表保存缓存项，并按访问时间排序，最近访问的在链表前端
// 2. 使用哈希表实现O(1)时间复杂度的随机查找
// 3. 每次访问或添加缓存项时，将其移至链表前端
// 4. 当需要淘汰时，移除链表尾部的项（最久未使用）
//
// 注意：该实现不是并发安全的。如需在并发环境使用，请添加适当的同步机制。
//
// 典型用途：
// - 数据库查询缓存
// - HTTP接口响应缓存
// - 热点文件内容缓存
// - 需要内存占用控制的任何缓存场景
package lru // LRU缓存包

import "container/list" // 导入Go标准库中的双向链表包

// Cache 是一个LRU（最近最少使用）缓存结构。注意：它不是并发安全的。
type Cache struct {
	maxBytes int64                    // 缓存的最大内存占用（字节）
	nbytes   int64                    // 当前缓存已使用的内存（字节）
	ll       *list.List               // 双向链表，用于维护缓存项的访问顺序
	cache    map[string]*list.Element // 字符串到链表节点的映射，用于O(1)时间复杂度查找缓存项
	// 可选的回调函数，当缓存项被清除时调用
	OnEvicted func(key string, value Value)
}

// entry 是存储在双向链表中的缓存项
type entry struct {
	key   string // 缓存项的键
	value Value  // 缓存项的值 **任何一个实现了Len()方法的类型**
}

// Value 接口用于计算值所占用的字节数
type Value interface {
	Len() int // 返回值所占用的字节数
}

// New 是Cache的构造函数
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,                            // 设置最大内存限制
		ll:        list.New(),                          // 初始化双向链表
		cache:     make(map[string]*list.Element, 100), // 初始化哈希表
		OnEvicted: onEvicted,                           // 设置回调函数
	}
}

// Add 向缓存中添加一个值
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		// 如果键已存在，更新对应节点的值
		c.ll.MoveToFront(ele)                                  // 将节点移到链表前端（表示最近访问）
		kv := ele.Value.(*entry)                               // 获取节点中存储的entry
		c.nbytes += int64(value.Len()) - int64(kv.value.Len()) // 更新内存占用（新值大小 - 旧值大小）
		kv.value = value                                       // 更新值
	} else {
		// 如果键不存在，创建新节点
		ele := c.ll.PushFront(&entry{key, value})        // 在链表前端添加新节点
		c.cache[key] = ele                               // 在哈希表中记录键到节点的映射
		c.nbytes += int64(len(key)) + int64(value.Len()) // 更新内存占用（键大小 + 值大小）
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		// 如果超过最大内存限制，移除最久未使用的节点
		c.RemoveOldest()
	}
}

// Get 查找键对应的值
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		// 如果键存在
		c.ll.MoveToFront(ele)    // 将节点移到链表前端（表示最近访问）
		kv := ele.Value.(*entry) // 获取节点中存储的entry
		return kv.value, true    // 返回值和true
	}
	return // 如果键不存在，返回零值和false
}

// RemoveOldest 移除最久未使用的缓存项
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back() // 获取链表尾部节点（最久未使用的）
	if ele != nil {
		c.ll.Remove(ele)                                       // 从链表中删除该节点
		kv := ele.Value.(*entry)                               // 获取节点中存储的entry
		delete(c.cache, kv.key)                                // 从哈希表中删除对应的键值对
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len()) // 更新内存占用
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value) // 如果设置了回调函数，调用它
		}
	}
}

// Len 返回缓存中的元素个数
func (c *Cache) Len() int {
	return c.ll.Len() // 返回链表长度
}
