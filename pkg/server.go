package kache

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/falldio/Kache/pkg/config"
	"github.com/falldio/Kache/pkg/consistenthash"
	pb "github.com/falldio/Kache/pkg/proto"
	"github.com/falldio/Kache/pkg/registry"
	log "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedKacheServer
	self    string // address:port
	mu      sync.Mutex
	peers   *consistenthash.Map
	running bool
	stopCh  chan error
	clients map[string]*Client
}

func NewServer(self string) *Server {
	return &Server{
		self: self,
	}
}

func (s *Server) Get(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	group, key := in.GetGroup(), in.GetKey()
	resp := &pb.Response{}

	log.Printf("[%s] Receives RPC Get request: %s/%s", s.self, group, key)
	if key == "" {
		return resp, fmt.Errorf("key required")
	}
	g := GetGroup(group)
	if g == nil {
		return resp, fmt.Errorf("group not found")
	}
	view, err := g.Get(key)
	if err != nil {
		return resp, err
	}
	resp.Value = view.ByteSlice()

	// update hot cache
	go func() {
		err = s.Update(group, key, resp.Value)
		if err != nil {
			log.Errorf("[%s] Updating %s/%s: %v", s.self, group, key, err)
		}
	}()
	return resp, nil
}

func (s *Server) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server already started")
	}
	s.running = true
	s.stopCh = make(chan error)

	port := strings.Split(s.self, ":")[1]
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("starting to listen on %s: %w", port, err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterKacheServer(grpcServer, s)

	// register service to etcd
	go func() {
		err := registry.Register("kache", s.self, s.stopCh)
		if err != nil {
			log.Fatalf(err.Error())
		}
		log.Printf("[%s] Revoke service and close tcp socket", s.self)
	}()

	s.mu.Unlock()

	if err := grpcServer.Serve(ln); s.running && err != nil {
		return fmt.Errorf("starting to serve: %w", err)
	}
	return nil
}

func (s *Server) SetPeers(peersAddr ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.peers == nil {
		s.peers = consistenthash.New(config.Config.DefaultReplicas, nil)
	}
	s.peers.Add(peersAddr...)
	s.clients = map[string]*Client{}
	for _, peerAddr := range peersAddr {
		if !validPeerAddr(peerAddr) {
			panic(fmt.Sprintf("[peer %s] invalid addr\n", peerAddr))
		}
		service := fmt.Sprintf("kache/%s", peerAddr)
		s.clients[peerAddr] = NewClient(service)
	}
}

// whether addr is in the format of x.x.x.x:port
func validPeerAddr(addr string) bool {
	l1 := strings.Split(addr, ":")
	if len(l1) != 2 {
		return false
	}
	l2 := strings.Split(l1[0], ".")
	if l1[0] != "localhost" && len(l2) != 4 {
		return false
	}
	return true
}

func (s *Server) PickPeer(key string) (PeerGetter, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	peerAddr := s.peers.Get(key)
	if peerAddr == s.self {
		log.Printf("[%s] this key is allocated to the local node\n", s.self)
		return nil, false
	}
	log.Printf("[%s] Pickk remote peer: %s\n", s.self, peerAddr)
	return s.clients[peerAddr], true
}

func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}
	s.stopCh <- nil
	s.running = false
	s.clients = nil
	s.peers = nil
}

func (s *Server) Update(group, key string, value []byte) error {
	cli, err := clientv3.New(registry.DefaultETCDConfig)
	if err != nil {
		return fmt.Errorf("creating etcd client: %w", err)
	}
	defer cli.Close()
	_, err = cli.Put(context.Background(), fmt.Sprintf("/%s/%s", group, key), string(value))
	if err != nil {
		return fmt.Errorf("updating %s/%s: %w", group, key, err)
	}
	return nil
}

var _ PeerPicker = (*Server)(nil)
