package gocache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultBasePath = "/_gocache/"

type HTTPPool struct {
	self     string // 本地地址和端口
	basePath string // 基础路径：节点间通讯地址的前缀
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
