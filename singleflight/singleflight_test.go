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
	results := make(map[string]string)
	var mu sync.Mutex // 添加互斥锁
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		v, err := g.Do("key1", func() (interface{}, error) {
			return "value1", nil
		})
		if err != nil {
			t.Errorf("Do error: %v", err)
			return
		}
		mu.Lock() // 加锁
		results["key1"] = v.(string)
		mu.Unlock() // 解锁
	}()

	go func() {
		defer wg.Done()
		v, err := g.Do("key2", func() (interface{}, error) {
			return "value2", nil
		})
		if err != nil {
			t.Errorf("Do error: %v", err)
			return
		}
		mu.Lock() // 加锁
		results["key2"] = v.(string)
		mu.Unlock() // 解锁
	}()

	go func() {
		defer wg.Done()
		v, err := g.Do("key3", func() (interface{}, error) {
			return "value3", nil
		})
		if err != nil {
			t.Errorf("Do error: %v", err)
			return
		}
		mu.Lock() // 加锁
		results["key3"] = v.(string)
		mu.Unlock() // 解锁
	}()

	wg.Wait()

	mu.Lock() // 加锁
	if results["key1"] != "value1" {
		t.Errorf("key1 got %v, want value1", results["key1"])
	}
	if results["key2"] != "value2" {
		t.Errorf("key2 got %v, want value2", results["key2"])
	}
	if results["key3"] != "value3" {
		t.Errorf("key3 got %v, want value3", results["key3"])
	}
	mu.Unlock() // 解锁
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
