package singleflight

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// 测试并发请求同一个key时，fn函数只执行一次
func TestDo(t *testing.T) {
	g := new(Group)
	counter := 0
	key := "test_key"

	// 模拟耗时操作
	fn := func() (any, error) {
		time.Sleep(100 * time.Millisecond)
		counter++
		return counter, nil
	}

	var wg sync.WaitGroup
	results := make([]int, 10)

	// 启动10个并发请求
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			val, _ := g.Do(key, fn)

			results[index] = val.(int)
		}(i)
	}

	wg.Wait()

	// 验证所有结果都相同，且counter只增加了一次
	for i := 0; i < 10; i++ {
		if results[i] != 1 {
			t.Errorf("结果不一致，期望1，得到%d", results[i])
		}
	}
	if counter != 1 {
		t.Errorf("函数执行次数错误，期望1，得到%d", counter)
	}
}

// 测试不同key的请求可以并发执行
func TestDoDifferentKeys(t *testing.T) {
	g := new(Group)
	counter := 0
	keys := []string{"key1", "key2", "key3"}

	fn := func() (any, error) {
		time.Sleep(100 * time.Millisecond)
		counter++
		return counter, nil
	}

	var wg sync.WaitGroup
	results := make(map[string]int)

	// 并发请求不同的key
	for _, key := range keys {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			val, _ := g.Do(k, fn)
			results[k] = val.(int)
		}(key)
	}

	wg.Wait()

	// 验证每个key的结果都不同
	seen := make(map[int]bool)
	for _, val := range results {
		if seen[val] {
			t.Errorf("不同key的结果重复: %d", val)
		}
		seen[val] = true
	}
	if counter != len(keys) {
		t.Errorf("函数执行次数错误，期望%d，得到%d", len(keys), counter)
	}
}

// 测试函数执行出错的情况
func TestDoError(t *testing.T) {
	g := new(Group)
	key := "error_key"
	expectedErr := fmt.Errorf("测试错误")

	fn := func() (any, error) {
		return nil, expectedErr
	}

	var wg sync.WaitGroup
	errors := make([]error, 5)

	// 并发请求
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			_, err := g.Do(key, fn)
			errors[index] = err
		}(i)
	}

	wg.Wait()

	// 验证所有请求都收到相同的错误
	for i := 0; i < 5; i++ {
		if errors[i] != expectedErr {
			t.Errorf("错误不一致，期望%v，得到%v", expectedErr, errors[i])
		}
	}
}

// 测试空key的情况
func TestDoEmptyKey(t *testing.T) {
	g := new(Group)
	fn := func() (any, error) {
		return "value", nil
	}

	_, err := g.Do("", fn)
	if err == nil {
		t.Error("期望空key返回错误，但未返回")
	}
}
