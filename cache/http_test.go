package gocachex_test

import (
	"fmt"
	gocachex "goCacheX/cache"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestHTTPPool(t *testing.T) {
	// 1. 创建缓存组，与原代码相同
	gocachex.NewGroup("scores", 2<<10, gocachex.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	// 2. 创建HTTP池，但不直接启动服务器
	peers := gocachex.NewHTTPPool("localhost:9999")

	// 3. 使用httptest创建测试服务器
	server := httptest.NewServer(peers)
	defer server.Close()

	// 4. 定义测试用例
	tests := []struct {
		name     string
		key      string
		wantBody string
		wantCode int
	}{
		{
			name:     "存在的键",
			key:      "Tom",
			wantBody: "630",
			wantCode: http.StatusOK,
		},
		{
			name:     "不存在的键",
			key:      "kkk",
			wantBody: "kkk not exist",
			wantCode: http.StatusInternalServerError,
		},
	}

	// 5. 执行测试用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.Printf("%v", server.URL)
			url := fmt.Sprintf("%s/_gocacheX/scores/%s", server.URL, tt.key)

			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("请求失败: %v", err)
			}
			defer resp.Body.Close()

			// 检查状态码
			if resp.StatusCode != tt.wantCode {
				t.Errorf("状态码不匹配: 期望 %d, 得到 %d", tt.wantCode, resp.StatusCode)
			}

			// 检查响应体
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("读取响应失败: %v", err)
			}

			bodyStr := string(body)
			if !strings.Contains(bodyStr, tt.wantBody) {
				t.Errorf("响应内容不匹配: 期望包含 %q, 得到 %q", tt.wantBody, bodyStr)
			}
		})
	}
}

func createGroup(name string) *gocachex.Group {
	return gocachex.NewGroup(name, 2<<10, gocachex.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func TestMultiNodeCache(t *testing.T) {
	// 初始化节点
	const numNodes = 3
	var servers []*httptest.Server
	var addrs []string

	// 创建每个节点的 server，并预先记录地址
	for i := 0; i < numNodes; i++ {
		server := httptest.NewServer(nil) // 占位，稍后设置 handler
		servers = append(servers, server)
		addrs = append(addrs, server.URL)
	}

	// 初始化每个节点的 group 和 handler
	for i, server := range servers {
		group := createGroup(fmt.Sprint(i))

		peers := gocachex.NewHTTPPool(server.URL)
		peers.Set(addrs...)
		group.RegisterPeers(peers)

		server.Config.Handler = peers
	}

	// 清理所有 server
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

	// 对每个节点发起请求
	for i, server := range servers {
		t.Run(fmt.Sprintf("node-%d", i+1), func(t *testing.T) {
			for _, tc := range testCases {
				t.Run(tc.key, func(t *testing.T) {
					url := fmt.Sprintf("%s/_gocacheX/%s/%s", server.URL, fmt.Sprint(i), tc.key)
					resp, err := http.Get(url)
					if err != nil {
						t.Fatalf("请求失败: %v", err)
					}
					defer resp.Body.Close()

					bytes, err := io.ReadAll(resp.Body)
					if err != nil {
						t.Fatalf("读取响应失败: %v", err)
					}

					got := strings.TrimSpace(string(bytes))
					if got != tc.expected {
						t.Errorf("节点 %d 期望值 %q，得到 %q", i+1, tc.expected, got)
					}
				})
			}
		})
	}
}
