// ```go
// 这个 ARC 实现的主要特点：

// 1. **数据结构**：
//    - T1：最近使用的条目列表
//    - T2：频繁使用的条目列表
//    - B1：最近使用的历史记录
//    - B2：频繁使用的历史记录
//    - 自适应参数 p：控制 T1 和 T2 的大小比例

// 2. **核心算法**：
//    - 自适应替换：根据访问模式动态调整 T1 和 T2 的大小
//    - 频率适应：通过 p 参数调整替换策略
//    - 历史记录：使用 B1 和 B2 记录被替换的条目

// 3. **性能优化**：
//    - 使用 `container/list` 实现高效的链表操作
//    - 使用互斥锁保证并发安全
//    - 使用 map 实现 O(1) 的查找

// 4. **特性**：
//    - 线程安全
//    - 自适应替换策略
//    - 支持并发操作
//    - 内存使用可控

// 使用示例：
// ```go
package lru

import (
	"fmt"
	"testing"
)

func TestARCBasic(t *testing.T) {
	arc := NewARC(3)

	// 测试基本操作
	arc.Put("key1", "value1")
	arc.Put("key2", "value2")
	arc.Put("key3", "value3")

	// 测试获取
	if v, ok := arc.Get("key1"); !ok || v != "value1" {
		t.Errorf("Get key1 failed, got %v, want value1", v)
	}

	// 测试更新
	arc.Put("key1", "newvalue1")
	if v, ok := arc.Get("key1"); !ok || v != "newvalue1" {
		t.Errorf("Update key1 failed, got %v, want newvalue1", v)
	}

	// 测试删除
	arc.Remove("key1")
	if _, ok := arc.Get("key1"); ok {
		t.Error("Remove key1 failed")
	}

	// 测试清空
	arc.Clear()
	if arc.Size() != 0 {
		t.Errorf("Clear failed, size is %d, want 0", arc.Size())
	}
}

func TestARCReplacement(t *testing.T) {
	arc := NewARC(3)

	// 测试替换策略
	arc.Put("key1", "value1")
	arc.Put("key2", "value2")
	arc.Put("key3", "value3")
	arc.Put("key4", "value4") // 应该触发替换

	// 验证 key1 是否被替换
	if _, ok := arc.Get("key1"); ok {
		t.Error("key1 should be replaced")
	}

	// 验证其他键是否存在
	if _, ok := arc.Get("key2"); !ok {
		t.Error("key2 should exist")
	}
	if _, ok := arc.Get("key3"); !ok {
		t.Error("key3 should exist")
	}
	if _, ok := arc.Get("key4"); !ok {
		t.Error("key4 should exist")
	}
}

func TestARCFrequency(t *testing.T) {
	arc := NewARC(3)

	// 测试频率适应
	arc.Put("key1", "value1")
	arc.Put("key2", "value2")
	arc.Put("key3", "value3")

	// 多次访问 key1
	for i := 0; i < 5; i++ {
		arc.Get("key1")
	}

	// 添加新键，应该替换 key2 或 key3
	arc.Put("key4", "value4")

	// 验证 key1 仍然存在
	if _, ok := arc.Get("key1"); !ok {
		t.Error("key1 should not be replaced")
	}
}

func TestARCCapacity(t *testing.T) {
	arc := NewARC(2)

	// 测试容量限制
	arc.Put("key1", "value1")
	arc.Put("key2", "value2")
	arc.Put("key3", "value3")

	if arc.Size() > 2 {
		t.Errorf("Size is %d, want 2", arc.Size())
	}
}

func TestARCConcurrent(t *testing.T) {
	arc := NewARC(100)
	done := make(chan bool)

	// 并发写入
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key%d", id*100+j)
				arc.Put(key, fmt.Sprintf("value%d", id*100+j))
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证结果
	if arc.Size() > 100 {
		t.Errorf("Size is %d, want <= 100", arc.Size())
	}
}

func BenchmarkARCPut(b *testing.B) {
	arc := NewARC(1000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		arc.Put(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}
}

func BenchmarkARCGet(b *testing.B) {
	arc := NewARC(1000)
	for i := 0; i < 1000; i++ {
		arc.Put(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arc.Get(fmt.Sprintf("key%d", i%1000))
	}
}
