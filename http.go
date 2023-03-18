package gocache

import (
	"fmt"
	"gocache/consistenthash"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_gocache/"
	defaultReplicas = 50
)

type HTTPPool struct {
	self        string                 // 本地地址和端口
	basePath    string                 // 基础路径：节点间通讯地址的前缀
	mu          sync.Mutex             // 保护节点和HttpGetters
	peers       *consistenthash.Map    // 一致性哈希算法根据具体的key选择节点
	httpGetters map[string]*httpGetter // 每个远端节点对应一个httpGetter
}

type httpGetter struct {
	baseURL string
}

// NewHTTPPool 初始化一个http节点
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log info with server name
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP 处理所有的http请求
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) { // 此请求不包含基础前缀
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// /<基础前缀>/<组名>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 { // 路径不合法
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0] // 取出组名
	key := parts[1]       // 取出key

	group := GetGroup(groupName) // 通过name获取group
	if group == nil {            // 此groupName不存在
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key) // 通过关键字获取字节视图
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 将缓存值作为http body进行响应
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

// Get 实现了PeerGetter接口
func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}

// Set 更新节点列表
func (p *HTTPPool) Set(peers ...string) {
	// consistenthash.Map非线程安全，所以这里要加锁
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	// 在hash环上添加真实节点和虚拟节点
	p.peers.Add(peers...)
	// 存储远端节点信息
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// PickPeer 根据key选择一个节点
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	// consistenthash.Map非线程安全，所以这里要加锁
	p.mu.Lock()
	defer p.mu.Unlock()
	// p.peers是个 哈希环，通过调用它的Get方法拿到远端节点.
	// 这里的 peer 是个地址
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

var _ PeerGetter = (*httpGetter)(nil)
