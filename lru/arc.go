package lru

import (
	"container/list"
	"sync"
	"time"
)

// ARC 实现自适应替换缓存
type ARC struct {
	// 缓存容量
	capacity int
	// 互斥锁
	mu sync.RWMutex
	// 最近使用的条目 (T1)
	t1 *list.List
	// 频繁使用的条目 (T2)
	t2 *list.List
	// 最近使用的条目的历史记录 (B1)
	b1 *list.List
	// 频繁使用的条目的历史记录 (B2)
	b2 *list.List
	// 缓存数据
	cache map[string]*list.Element
	// 当前大小
	size int
	// 自适应参数 p
	p int
	// 停止清理的通道
	stopCh chan struct{}
}

// arcEntry 表示缓存条目
type arcEntry struct {
	key   string
	value any
	// 用于区分 T1 和 T2 中的条目
	inT2 bool
	// 过期时间
	expireAt time.Time
}

// NewARC 创建一个新的 ARC 缓存
func NewARC(capacity int) *ARC {
	arc := &ARC{
		capacity: capacity,
		t1:       list.New(),
		t2:       list.New(),
		b1:       list.New(),
		b2:       list.New(),
		cache:    make(map[string]*list.Element),
		p:        0,
		stopCh:   make(chan struct{}),
	}
	// 启动清理协程
	go arc.cleanupLoop()
	return arc
}

// cleanupLoop 定期清理过期条目
func (arc *ARC) cleanupLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			arc.cleanup()
		case <-arc.stopCh:
			return
		}
	}
}

// cleanup 清理所有列表中的过期条目
func (arc *ARC) cleanup() {
	// 清理 T1
	arc.cleanupList(arc.t1)
	// 清理 T2
	arc.cleanupList(arc.t2)
	// 清理 B1
	arc.cleanupList(arc.b1)
	// 清理 B2
	arc.cleanupList(arc.b2)
}

// cleanupList 清理指定列表中的过期条目
func (arc *ARC) cleanupList(l *list.List) {
	now := time.Now()
	for e := l.Front(); e != nil; {
		next := e.Next()
		entry, ok := e.Value.(*arcEntry)
		if !ok {
			// 如果类型转换失败，跳过此元素
			e = next
			continue
		}
		if !entry.expireAt.IsZero() && now.After(entry.expireAt) {
			l.Remove(e)
			delete(arc.cache, entry.key)
			arc.size--
		}
		e = next
	}
}

// Put 添加或更新缓存值
func (arc *ARC) Put(key string, value interface{}) {
	arc.PutWithTTL(key, value, 0)
}

// PutWithTTL 添加或更新缓存值，带过期时间
func (arc *ARC) PutWithTTL(key string, value interface{}, ttl time.Duration) {
	arc.mu.Lock()
	defer arc.mu.Unlock()

	// 检查 TTL 是否有效
	if ttl < 0 {
		// 如果 TTL 为负，直接返回
		return
	}

	// 如果键已存在
	if ele, ok := arc.cache[key]; ok {
		// 更新值和过期时间
		entry := ele.Value.(*arcEntry)
		entry.value = value
		if ttl > 0 {
			entry.expireAt = time.Now().Add(ttl)
		} else {
			entry.expireAt = time.Time{}
		}
		// 如果元素在 T1 中
		if !entry.inT2 {
			// 从 T1 移动到 T2
			arc.t1.Remove(ele)
			entry.inT2 = true
			arc.t2.PushFront(ele)
		} else {
			// 如果元素在 T2 中，移动到 T2 的前面
			arc.t2.MoveToFront(ele)
		}
		return
	}

	// 创建新条目
	ent := &arcEntry{
		key:   key,
		value: value,
		inT2:  false,
	}
	if ttl > 0 {
		ent.expireAt = time.Now().Add(ttl)
	}

	// 如果缓存未满
	if arc.size < arc.capacity {
		ele := arc.t1.PushFront(ent)
		arc.cache[key] = ele
		arc.size++
		return
	}

	// 自适应替换
	arc.replace(ent)
}

// Get 获取缓存值
func (arc *ARC) Get(key string) (interface{}, bool) {
	arc.mu.Lock()
	defer arc.mu.Unlock()

	if ele, ok := arc.cache[key]; ok {
		entry := ele.Value.(*arcEntry)
		// 检查是否过期
		if !entry.expireAt.IsZero() && time.Now().After(entry.expireAt) {
			// 如果过期，删除条目
			if entry.inT2 {
				arc.t2.Remove(ele)
			} else {
				arc.t1.Remove(ele)
			}
			delete(arc.cache, key)
			arc.size--
			return nil, false
		}

		// 如果元素在 T1 中
		if !entry.inT2 {
			// 从 T1 移动到 T2
			arc.t1.Remove(ele)
			entry.inT2 = true
			arc.t2.PushFront(ele)
		} else {
			// 如果元素在 T2 中，移动到 T2 的前面
			arc.t2.MoveToFront(ele)
		}
		return entry.value, true
	}
	return nil, false
}

// Close 关闭缓存，停止清理协程
func (arc *ARC) Close() {
	close(arc.stopCh)
}

// replace 执行替换操作
func (arc *ARC) replace(ent *arcEntry) {
	// 如果 T1 和 T2 都为空，直接添加新条目
	if arc.t1.Len() == 0 && arc.t2.Len() == 0 {
		ele := arc.t1.PushFront(ent)
		arc.cache[ent.key] = ele
		return
	}

	var last *list.Element
	var lastEntry *arcEntry

	// 如果 T1 不为空且 (p > 0 或 B2 为空)
	if arc.t1.Len() > 0 && (arc.p > 0 || arc.b2.Len() == 0) {
		last = arc.t1.Back()
		if last == nil {
			return
		}

		var ok bool
		lastEntry, ok = last.Value.(*arcEntry)
		if !ok {
			return
		}

		arc.t1.Remove(last)
		// 将元素移动到 B1，并限制 B1 的大小
		arc.b1.PushFront(lastEntry)
		lastEntry.inT2 = false
		if arc.b1.Len() > arc.capacity {
			if old := arc.b1.Back(); old != nil {
				arc.b1.Remove(old)
			}
		}
		// 减小 p
		arc.p = max(0, arc.p-1)
	} else {
		last = arc.t2.Back()
		if last == nil {
			return
		}

		var ok bool
		lastEntry, ok = last.Value.(*arcEntry)
		if !ok {
			return
		}

		arc.t2.Remove(last)
		// 将元素移动到 B2，并限制 B2 的大小
		arc.b2.PushFront(lastEntry)
		lastEntry.inT2 = true
		if arc.b2.Len() > arc.capacity {
			if old := arc.b2.Back(); old != nil {
				arc.b2.Remove(old)
			}
		}
		// 增加 p
		arc.p = min(arc.capacity, arc.p+1)
	}

	// 删除缓存中的旧条目
	if lastEntry != nil {
		delete(arc.cache, lastEntry.key)
	}

	// 添加新条目到 T1
	ele := arc.t1.PushFront(ent)
	arc.cache[ent.key] = ele
}

// Remove 删除缓存值
func (arc *ARC) Remove(key string) {
	arc.mu.Lock()
	defer arc.mu.Unlock()

	if ele, ok := arc.cache[key]; ok {
		if ele.Value.(*arcEntry).inT2 {
			arc.t2.Remove(ele)
		} else {
			arc.t1.Remove(ele)
		}
		delete(arc.cache, key)
		arc.size--
	}
}

// Clear 清空缓存
func (arc *ARC) Clear() {
	arc.mu.Lock()
	defer arc.mu.Unlock()

	arc.t1.Init()
	arc.t2.Init()
	arc.b1.Init()
	arc.b2.Init()
	arc.cache = make(map[string]*list.Element)
	arc.size = 0
	arc.p = 0
}

// Size 返回当前缓存大小
func (arc *ARC) Size() int {
	arc.mu.RLock()
	defer arc.mu.RUnlock()
	return arc.size
}

// Capacity 返回缓存容量
func (arc *ARC) Capacity() int {
	return arc.capacity
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
