// Package consistenthash 实现一致性哈希算法
// 一致性哈希算法用于将key分配到节点上，并在节点变化时尽量减少key的重新分配
package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 定义哈希函数类型
type Hash func(data []byte) uint32

// Map 是一致性哈希算法的主要数据结构
type Map struct {
	hash      Hash           // 哈希函数
	nreplicas int            // 虚拟节点倍数
	keys      []int          // 哈希环上的已排序节点哈希值
	mapping   map[int]string // 节点哈希值到节点名的映射
}

// NewMap 创建一个Map实例
// nreplicas: 虚拟节点倍数，每个真实节点对应多少个虚拟节点
// hashfunc: 自定义的哈希函数，如果为nil则使用crc32.ChecksumIEEE
func NewMap(nreplicas int, hashfunc Hash) *Map {
	m := &Map{
		hash:      hashfunc,
		nreplicas: nreplicas,
		mapping:   make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 添加节点到哈希环
// 为每个节点创建nreplicas个虚拟节点
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.nreplicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.mapping[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Get 根据key选择节点
// 返回哈希环上顺时针方向最近的节点
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	// 二分查找，找到第一个大于等于hash的节点
	index := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	// 如果没找到，或者找到的位置超出切片范围，则环绕到第一个节点
	return m.mapping[m.keys[index%len(m.keys)]]
}
