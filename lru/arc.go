package lru

import (
	"container/list"
	"sync"
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
}

// arcEntry 表示缓存条目
type arcEntry struct {
	key   string
	value any
	// 用于区分 T1 和 T2 中的条目
	inT2 bool
}

// NewARC 创建一个新的 ARC 缓存
func NewARC(capacity int) *ARC {
	return &ARC{
		capacity: capacity,
		t1:       list.New(),
		t2:       list.New(),
		b1:       list.New(),
		b2:       list.New(),
		cache:    make(map[string]*list.Element),
		p:        0,
	}
}

// Get 获取缓存值
func (arc *ARC) Get(key string) (interface{}, bool) {
	arc.mu.Lock()
	defer arc.mu.Unlock()

	if ele, ok := arc.cache[key]; ok {
		// 如果元素在 T1 中
		if !ele.Value.(*arcEntry).inT2 {
			// 从 T1 移动到 T2
			arc.t1.Remove(ele)
			ele.Value.(*arcEntry).inT2 = true
			arc.t2.PushFront(ele)
		} else {
			// 如果元素在 T2 中，移动到 T2 的前面
			arc.t2.MoveToFront(ele)
		}
		return ele.Value.(*arcEntry).value, true
	}
	return nil, false
}

// Put 添加或更新缓存值
func (arc *ARC) Put(key string, value interface{}) {
	arc.mu.Lock()
	defer arc.mu.Unlock()

	// 如果键已存在
	if ele, ok := arc.cache[key]; ok {
		// 更新值
		ele.Value.(*arcEntry).value = value
		// 如果元素在 T1 中
		if !ele.Value.(*arcEntry).inT2 {
			// 从 T1 移动到 T2
			arc.t1.Remove(ele)
			ele.Value.(*arcEntry).inT2 = true
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

	// 如果缓存未满
	if arc.size < arc.capacity {
		arc.t1.PushFront(ent)
		arc.cache[key] = arc.t1.Front()
		arc.size++
		return
	}

	// 自适应替换
	arc.replace(ent)
}

// replace 执行替换操作
func (arc *ARC) replace(ent *arcEntry) {
	// 如果 T1 和 T2 都为空，直接添加新条目
	if arc.t1.Len() == 0 && arc.t2.Len() == 0 {
		arc.t1.PushFront(ent)
		arc.cache[ent.key] = arc.t1.Front()
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
	arc.t1.PushFront(ent)
	arc.cache[ent.key] = arc.t1.Front()
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
