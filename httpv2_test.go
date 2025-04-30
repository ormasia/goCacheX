package main

import (
	"fmt"
	cache "goCacheX/cache"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 模拟数据库
var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

// 创建缓存组
func createGroup() *cache.Group {
	return cache.NewGroup("scores", 2<<10, cache.GetterFunc(
		func(key string) ([]byte, error) {
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

// 测试多节点缓存
func TestMultiNodeCache(t *testing.T) {
	// 定义多个测试服务器
	nodes := []struct {
		addr string
		port int
	}{
		{"http://localhost", 8001},
		{"http://localhost", 8002},
		{"http://localhost", 8003},
	}

	// 创建并启动所有节点
	var servers []*httptest.Server
	var addrs []string
	// var groups []*cache.Group

	for _, node := range nodes {
		addr := fmt.Sprintf("%s:%d", node.addr, node.port)
		addrs = append(addrs, addr)

		// 为每个节点创建独立的缓存组
		gee := createGroup()
		// groups = append(groups, gee)

		peers := cache.NewHTTPPool(addr)
		peers.Set(addrs...)
		gee.RegisterPeers(peers)

		// 启动服务器
		server := httptest.NewServer(peers)
		servers = append(servers, server)
	}

	// 确保所有服务器都被关闭
	defer func() {
		for _, server := range servers {
			server.Close()
		}
	}()

	// 测试用例
	testCases := []struct {
		key      string
		expected string
	}{
		{"Tom", "630"},
		{"Jack", "589"},
		{"Sam", "567"},
	}

	// 测试每个节点
	for i, server := range servers {
		t.Run(fmt.Sprintf("node-%d", i+1), func(t *testing.T) {
			for _, tc := range testCases {
				t.Run(tc.key, func(t *testing.T) {
					// 发送请求到当前节点
					resp, err := http.Get(fmt.Sprintf("%s/_gocacheX/scores/%s", server.URL, tc.key))
					if err != nil {
						t.Fatalf("请求失败: %v", err)
					}
					defer resp.Body.Close()

					bytes, err := io.ReadAll(resp.Body)
					if err != nil {
						t.Fatalf("读取响应失败: %v", err)
					}

					if string(bytes) != tc.expected {
						t.Errorf("节点 %d 期望值 %s，得到 %s", i+1, tc.expected, string(bytes))
					}
				})
			}
		})
	}
}

// 测试节点故障
func TestNodeFailure(t *testing.T) {
	// 创建两个节点
	nodes := []struct {
		addr string
		port int
	}{
		{"http://localhost", 8001},
		{"http://localhost", 8002},
	}

	var servers []*httptest.Server
	var addrs []string
	// var groups []*cache.Group

	// 只启动第一个节点
	for i, node := range nodes {
		addr := fmt.Sprintf("%s:%d", node.addr, node.port)
		addrs = append(addrs, addr)

		gee := createGroup()
		// groups = append(groups, gee)

		peers := cache.NewHTTPPool(addr)
		peers.Set(addrs...)
		gee.RegisterPeers(peers)

		// 只启动第一个节点
		if i == 0 {
			server := httptest.NewServer(peers)
			servers = append(servers, server)
		}
	}

	defer func() {
		for _, server := range servers {
			server.Close()
		}
	}()

	// 测试数据
	testCases := []struct {
		key      string
		expected string
	}{
		{"Tom", "630"},
		{"Jack", "589"},
	}

	// 测试第一个节点（正常工作的节点）
	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			resp, err := http.Get(fmt.Sprintf("%s/_gocacheX/scores/%s", servers[0].URL, tc.key))
			if err != nil {
				t.Fatalf("请求失败: %v", err)
			}
			defer resp.Body.Close()

			bytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("读取响应失败: %v", err)
			}

			if string(bytes) != tc.expected {
				t.Errorf("期望值 %s，得到 %s", tc.expected, string(bytes))
			}
		})
	}
}
