package geecache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// 提供被其他节点访问的能力(基于http)
const defaultBasePath = "/_geecache/"

type HTTPPool struct {
	self     string // 记录自己的地址 主机ip和端口
	basePath string // 作为节点间通讯地址的前缀，默认是/_geecache/  localhost:8080/_geecache/get...
}

func NewHttPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log log info
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v))
}

// ServeHttp handle all http requests
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 判断请求前缀是否为默认前缀
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// basePath/groupname/key
	// basePath之后的path按/进行分割 获取key
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// 得到groupname 和 key
	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())

}
