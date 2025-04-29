// consistenthash_test.go - 一致性哈希算法的测试文件
package consistenthash

import (
	"strconv"
	"testing"
)

// TestHashing 测试一致性哈希的基本功能
// 1. 测试正确的节点选择
// 2. 测试添加新节点后的重新分配
func TestHashing(t *testing.T) {
	// 创建一致性哈希实例，使用自定义哈希函数
	// 这里使用简单的数字转换作为哈希函数，便于测试结果的可预测性
	hash := NewMap(3, func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	})

	// 添加3个节点："6", "4", "2"
	// 由于虚拟节点倍数为3，所以每个节点会产生3个虚拟节点
	// 根据上面定义的哈希函数，虚拟节点的哈希值分别是：
	// 2, 4, 6, 12, 14, 16, 22, 24, 26
	hash.Add("6", "4", "2")

	// 测试用例：key到节点的映射
	testCases := map[string]string{
		"2":  "2", // 哈希值2，应映射到节点"2"
		"11": "2", // 哈希值11，顺时针找到节点"2"(哈希值12)
		"23": "4", // 哈希值23，顺时针找到节点"4"(哈希值24)
		"27": "2", // 哈希值27，环绕回到节点"2"(哈希值2)
	}

	// 验证每个key是否映射到正确的节点
	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("请求键 %s, 应返回节点 %s", k, v)
		}
	}

	// 添加新节点："8"
	// 新增虚拟节点的哈希值为：8, 18, 28
	hash.Add("8")

	// 更新测试用例：重新分配后，键"27"应该映射到新节点"8"
	testCases["27"] = "8"

	// 再次验证所有测试用例
	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("请求键 %s, 应返回节点 %s", k, v)
		}
	}
}
