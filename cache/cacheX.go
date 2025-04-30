/*
	    cacheX构建一个分布式缓存框架，
		- 用Group隔离不同类型的缓存数据
		- 用cache封装并发安全的缓存访问
		- 用lru实现缓存淘汰
		- 用ByteView封装缓存数据，确保数据安全
		- 用Getter定义缓存未命中时的回调函数
*/
package gocachex

import (
	"fmt"
	pb "goCacheX/gocacheXpb"
	"goCacheX/singleflight"
	"log"
	"sync"
)

// Group 是缓存的命名空间，每个Group拥有一个唯一的名称
// 代表一个独立的缓存空间，管理特定类型的缓存数据
type Group struct {
	name      string // 缓存命名空间的名称
	getter    Getter // 缓存未命中时获取源数据的回调函数
	mainCache cache  // 并发安全的主缓存，存储实际的缓存数据

	peers  PeerPicker          // 通过一致性哈希选择节点
	loader *singleflight.Group // 防止缓存击穿
}

// Getter 定义了当缓存未命中时获取源数据的接口
// 实现此接口的对象负责从数据源获取原始数据
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 是一个实现了Getter接口的函数类型
// 允许将普通函数转换为Getter接口使用
type GetterFunc func(key string) ([]byte, error)

// Get 调用函数本身，实现Getter接口
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group) // 全局变量，存储所有Group实例
)

// NewGroup 创建一个新的缓存分组实例
// name: 分组名称，cacheBytes: 缓存最大内存限制，getter: 缓存未命中时的回调
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

// GetGroup 根据名称获取对应的缓存分组
func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	g := groups[name]
	return g
}

// Get 从缓存获取键对应的值，如果缓存中不存在，则调用load方法加载
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	bytes, ok := g.mainCache.get(key)
	if ok {
		log.Println("[GeeCache] hit")
		return bytes, nil
	}
	return g.load(key)
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// load 加载键对应的值，可以从本地或远程获取
func (g *Group) load(key string) (value ByteView, err error) { //返回值变量在函数开始时就已声明和初始化可以直接在函数体内使用这些变量不需要显式 return 具体的值，可以直接 return适合需要多次修改返回值的情况

	view, err := g.loader.Do(key, func() (any, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return view.(ByteView), nil
	}
	return ByteView{}, err
}

// getLocally 从本地数据源获取原始数据，转换为ByteView并添加到缓存
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	// 使用cloneBytes创建原始数据的深拷贝的原因：？？
	// 1. 防止外部修改：即使原始bytes在外部被修改，也不会影响缓存中的数据
	// 2. 保持ByteView的只读特性：确保缓存数据的不可变性，增强缓存的稳定性
	// 3. 内存所有权清晰：缓存系统完全控制这部分内存，不依赖外部代码的内存管理
	// 4. 并发安全考虑：不可变数据更适合在并发环境中使用，减少潜在的竞态条件
	// 虽然有轻微性能开销，但换来更好的数据安全性和系统稳定性
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

// populateCache 将键值对添加到缓存
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}
