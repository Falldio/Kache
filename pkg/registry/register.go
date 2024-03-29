package registry

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
)

var (
	DefaultETCDConfig = clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	}
)

// add kv to etcd under lease mode
func etcdAdd(c *clientv3.Client, lid clientv3.LeaseID, service string, addr string) error {
	em, err := endpoints.NewManager(c, service)
	if err != nil {
		return err
	}
	return em.AddEndpoint(c.Ctx(), service+"/"+addr, endpoints.Endpoint{Addr: addr}, clientv3.WithLease(lid))
}

// register a service to etcd
func Register(service string, addr string, stop chan error) error {
	cli, err := clientv3.New(DefaultETCDConfig)
	if err != nil {
		return fmt.Errorf("creating etcd client: %w", err)
	}
	defer cli.Close()

	resp, err := cli.Grant(context.Background(), 5)
	if err != nil {
		return fmt.Errorf("creating lease: %w", err)
	}
	leaseId := resp.ID

	err = etcdAdd(cli, leaseId, service, addr)
	if err != nil {
		return fmt.Errorf("adding etcd record: %w", err)
	}
	ch, err := cli.KeepAlive(context.Background(), leaseId)
	if err != nil {
		return fmt.Errorf("setting keepalive: %w", err)
	}
	log.Printf("[%s] servise registered\n", addr)
	for {
		select {
		case err := <-stop:
			if err != nil {
				log.Println(err)
			}
			return err
		case <-cli.Ctx().Done():
			log.Println("service closed")
			return nil
		case _, ok := <-ch:
			if !ok {
				log.Println("keep alive channel closed")
				_, err := cli.Revoke(context.Background(), leaseId)
				return err
			}
		}
	}
}
