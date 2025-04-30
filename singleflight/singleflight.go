package singleflight

import (
	"fmt"
	"sync"
)

type call struct {
	wg  sync.WaitGroup
	val any
	err error
}

type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

func (g *Group) Do(key string, fn func() (any, error)) (any, error) {
	g.mu.Lock()
	// defer g.mu.Unlock() fatal error: sync: unlock of unlocked mutex
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if key == "" {
		g.mu.Unlock()
		return nil, fmt.Errorf("key is empty")
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	c := new(call)
	c.wg.Add(1)
	g.m[key] = c //证明已经在执行
	g.mu.Unlock()

	// 在后台执行实际函数（非阻塞）
	go func() {
		defer c.wg.Done() // 确保无论是否panic都标记完成
		c.val, c.err = fn()
	}()

	c.wg.Wait() // 等待后台函数完成

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
