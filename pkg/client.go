package kache

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/falldio/Kache/pkg/proto"
	"github.com/falldio/Kache/pkg/registry"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Client struct {
	// service name: kache/ip:addr
	name string
}

func (c *Client) Get(group string, key string) ([]byte, error) {
	cli, err := clientv3.New(registry.DefaultETCDConfig)
	if err != nil {
		return nil, fmt.Errorf("creating etcd client: %w", err)
	}
	defer cli.Close()

	conn, err := registry.ETCDDial(cli, c.name)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	grpcClient := pb.NewKacheClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := grpcClient.Get(ctx, &pb.Request{
		Group: group,
		Key:   key,
	})
	if err != nil {
		return nil, fmt.Errorf("getting %s/%s from peer %s: %w", group, key, c.name, err)
	}

	return resp.GetValue(), nil
}

func NewClient(service string) *Client {
	return &Client{name: service}
}

func (c *Client) Watch(group string, key string, onUpdated func([]byte)) {
	cli, err := clientv3.New(registry.DefaultETCDConfig)
	if err != nil {
		log.Fatalf("creating etcd client: %v", err)
	}
	defer cli.Close()
	rch := cli.Watch(context.Background(), fmt.Sprintf("/%s/%s", group, key))
	for wresp := range rch {
		for _, ev := range wresp.Events {
			onUpdated(ev.Kv.Value)
		}
	}
}

var _ PeerGetter = (*Client)(nil)
