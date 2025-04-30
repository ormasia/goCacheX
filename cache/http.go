package gocachex

import (
	"fmt"
	"goCacheX/consistenthash"
	pb "goCacheX/gocacheXpb"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"google.golang.org/protobuf/proto"
)

// 定义常量
const (
	defaultBasePath = "/_gocacheX/" // 默认的HTTP请求路径前缀
	defaultReplicas = 50            // 一致性哈希的默认虚拟节点数
)

// HTTPPool 实现了 PeerPicker 接口，用于管理HTTP节点池
type HTTPPool struct {
	self        string                 // 当前节点的URL，例如 "https://example.net:8000"
	basePath    string                 // HTTP请求的基础路径
	mu          sync.Mutex             // 互斥锁，保护并发访问
	peers       *consistenthash.Map    // 一致性哈希映射，用于节点选择
	httpGetters map[string]*httpGetter // 节点到httpGetter的映射，用于向其他节点发送HTTP请求获取缓存数据
}

// NewHTTPPool 初始化一个HTTP节点池
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log 记录服务器日志，包含服务器名称
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP 处理所有HTTP请求
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 检查请求路径是否以basePath开头
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)

	// 解析请求路径：/<basepath>/<groupname>/<key>
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	// 获取对应的缓存组
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	// 从缓存组获取数据
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 将数据序列化为protobuf格式
	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 设置响应头并返回数据
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}

// Set 设置节点池中的节点
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 初始化一致性哈希映射
	p.peers = consistenthash.NewMap(defaultReplicas, nil)
	p.peers.Add(peers...)

	// 为每个节点创建httpGetter
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		// baseURL格式：<peer>_<basepath>/<groupname>/<key>
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// PickPeer 根据key选择一个节点
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 通过一致性哈希选择节点，并防止选择自身
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

// 确保HTTPPool实现了PeerPicker接口
var _ PeerPicker = (*HTTPPool)(nil)

// httpGetter 实现了PeerGetter接口，用于从其他节点获取数据
type httpGetter struct {
	baseURL string // 基础URL，用于构建完整的请求URL
}

// Get 通过HTTP请求获取指定group的key数据
func (h *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	// 构建请求URL
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(in.GetGroup()), // 对group名称进行URL编码
		url.QueryEscape(in.GetKey()),   // 对key进行URL编码
	)

	// 发送GET请求
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// 检查响应状态码
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	// 读取响应体
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}

	// 解析protobuf响应
	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}

	return nil
}

// 确保httpGetter实现了PeerGetter接口
var _ PeerGetter = (*httpGetter)(nil)
