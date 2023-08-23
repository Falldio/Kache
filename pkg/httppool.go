package kache

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/falldio/Kache/pkg/consistenthash"
	pb "github.com/falldio/Kache/pkg/proto/gen"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

const (
	DefaultBasePath = "/_kache/"
	DefaultReplicas = 50
)

type HTTPPool struct {
	self        string // address:port
	basePath    string
	mu          sync.Mutex
	peers       *consistenthash.Map
	httpGetters map[string]*httpGetter
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: DefaultBasePath,
	}
}

func (p *HTTPPool) Log(format string, v ...any) {
	log.Printf("[Server %s]%s", p.self, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(c *gin.Context) {
	if !strings.HasPrefix(c.Request.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + c.Request.URL.Path)
	}
	groupName := c.Query("group")
	key := c.Query("key")
	p.Log("%s %s", groupName, key)

	group := GetGroup(groupName)
	if group == nil {
		c.AbortWithError(http.StatusNotFound, fmt.Errorf("no such group: %s", groupName))
		return
	}
	view, err := group.Get(key)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}
	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("message marshaling failed"))
	}
	c.Data(http.StatusOK, "application/octet-stream", body)
}

func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(DefaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)
