package services

import (
	"errors"
	log "github.com/GameGophers/nsq-logger"
	"github.com/coreos/go-etcd/etcd"
	"google.golang.org/grpc"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

import (
	"services/proto"
)

var (
	ERROR_SERVICE_NOT_AVAILABLE = errors.New("service not available")
)

const (
	DEFAULT_ETCD         = "http://127.0.0.1:2379"
	DEFAULT_SERVICE_PATH = "/backends"
	DEFAULT_DIAL_TIMEOUT = 10 * time.Second
	RETRY_DELAY          = 10 * time.Second
)

type client struct {
	key  string
	conn *grpc.ClientConn
}

type service_pool struct {
	snowflakes     []client
	snowflakes_idx int
	client_pool    sync.Pool
	sync.Mutex
}

var (
	_default_pool service_pool
)

func init() {
	_default_pool.init()
	_default_pool.connect_all(DEFAULT_SERVICE_PATH)
	go _default_pool.watcher()
}

func (p *service_pool) init() {
	// etcd client
	machines := []string{DEFAULT_ETCD}
	if env := os.Getenv("ETCD_HOST"); env != "" {
		machines = strings.Split(env, ";")
	}
	p.client_pool.New = func() interface{} {
		return etcd.NewClient(machines)
	}
}

// connect to all services
func (p *service_pool) connect_all(directory string) {
	client := p.client_pool.Get().(*etcd.Client)
	defer func() {
		p.client_pool.Put(client)
	}()

	// get the keys under directory
	log.Info("connecting services under:", directory)
	resp, err := client.Get(directory, true, true)
	if err != nil {
		log.Error(err)
		return
	}

	// validation check
	if !resp.Node.Dir {
		log.Error("not a directory")
		return
	}

	for _, node := range resp.Node.Nodes {
		if node.Dir { // service directory
			for _, service := range node.Nodes {
				log.Tracef("add node: %v %v", service.Key, service.Value)
				p.add_node(service.Key, service.Value)
			}
		} else {
			log.Warning("malformed service directory:", node.Key)
		}
	}
	log.Trace("connected to all services")
}

func (p *service_pool) delete_node(key string) {
	p.Lock()
	defer p.Unlock()
	switch filepath.Dir(key) {
	case DEFAULT_SERVICE_PATH + "/snowflake":
		for k := range p.snowflakes {
			if p.snowflakes[k].key == key { // deletion
				p.snowflakes[k].conn.Close()
				p.snowflakes = append(p.snowflakes[:k], p.snowflakes[k+1:]...)
				return
			}
		}
	}
}

func (p *service_pool) add_node(key, value string) {
	p.Lock()
	defer p.Unlock()
	switch filepath.Dir(key) {
	case DEFAULT_SERVICE_PATH + "/snowflake":
		if conn, err := grpc.Dial(value, grpc.WithTimeout(DEFAULT_DIAL_TIMEOUT)); err == nil {
			p.snowflakes = append(p.snowflakes, client{key, conn})
		} else {
			log.Errorf("did not connect: %v %v err: %v", key, value, err)
		}
	default:
		log.Warningf("service not recongized: %v %v", key, value)
	}
}

// watcher for data change in etcd directory
func (p *service_pool) watcher() {
	client := p.client_pool.Get().(*etcd.Client)
	defer func() {
		p.client_pool.Put(client)
	}()

	for {
		ch := make(chan *etcd.Response, 10)
		go func() {
			for {
				if resp, ok := <-ch; ok {
					if resp.Node.Dir {
						continue
					}
					key, value := resp.Node.Key, resp.Node.Value
					if value == "" {
						log.Tracef("node delete: %v", key)
						p.delete_node(key)
					} else {
						log.Tracef("node add: %v %v", key, value)
						p.add_node(key, value)
					}
				} else {
					return
				}
			}
		}()

		_, err := client.Watch(DEFAULT_SERVICE_PATH, 0, true, ch, nil)
		if err != nil {
			log.Critical(err)
		}
		<-time.After(RETRY_DELAY)
	}
}

func (p *service_pool) get_snowflake() (proto.SnowflakeServiceClient, error) {
	p.Lock()
	defer p.Unlock()
	if len(p.snowflakes) == 0 {
		return nil, ERROR_SERVICE_NOT_AVAILABLE
	}

	p.snowflakes_idx++
	return proto.NewSnowflakeServiceClient(p.snowflakes[p.snowflakes_idx%len(p.snowflakes)].conn), nil
}

// wrappers
func GetSnowflake() (proto.SnowflakeServiceClient, error) {
	return _default_pool.get_snowflake()
}
