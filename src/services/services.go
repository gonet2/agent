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
	"sync/atomic"
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

type service struct {
	clients []client
	idx     uint32
}

type service_pool struct {
	services    map[string]*service
	client_pool sync.Pool
	sync.RWMutex
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

	p.services = make(map[string]*service)
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
				p.add_service(service.Key, service.Value)
			}
		} else {
			log.Warning("malformed service directory:", node.Key)
		}
	}
	log.Info("services add complete")
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
						p.remove_service(key)
					} else {
						log.Tracef("node add: %v %v", key, value)
						p.add_service(key, value)
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

// add a service
func (p *service_pool) add_service(key, value string) {
	p.Lock()
	defer p.Unlock()
	service_name := filepath.Dir(key)
	if p.services[service_name] == nil {
		p.services[service_name] = &service{}
		log.Tracef("new service type: %v", service_name)
	}
	service := p.services[service_name]

	if conn, err := grpc.Dial(value, grpc.WithTimeout(DEFAULT_DIAL_TIMEOUT)); err == nil {
		service.clients = append(service.clients, client{key, conn})
		log.Tracef("service added: %v -- %v", key, value)
	} else {
		log.Errorf("did not connect: %v -- %v err: %v", key, value, err)
	}
}

// remove a service
func (p *service_pool) remove_service(key string) {
	p.Lock()
	defer p.Unlock()
	service_name := filepath.Dir(key)
	service := p.services[service_name]
	if service == nil {
		log.Tracef("no such service %v", service_name)
		return
	}

	for k := range service.clients {
		if service.clients[k].key == key { // deletion
			service.clients[k].conn.Close()
			service.clients = append(service.clients[:k], service.clients[k+1:]...)
			log.Tracef("service removed %v", key)
			return
		}
	}
}

// provide a specific key for a service
// service must be stored like /backends/xxx_service/xxx_id
func (p *service_pool) get_service_with_id(service_name, id string) (interface{}, error) {
	p.RLock()
	defer p.RUnlock()
	name := DEFAULT_SERVICE_PATH + "/" + service_name
	service := p.services[name]
	if service == nil {
		return nil, ERROR_SERVICE_NOT_AVAILABLE
	}

	if len(service.clients) == 0 {
		return nil, ERROR_SERVICE_NOT_AVAILABLE
	}

	var conn *grpc.ClientConn
	fullpath := name + "/" + id
	for k := range service.clients {
		if service.clients[k].key == fullpath {
			conn = service.clients[k].conn
			break
		}
	}

	if conn == nil {
		return nil, ERROR_SERVICE_NOT_AVAILABLE
	}

	// add wrappers here ...
	switch ServiceType(name) {
	case SERVICE_SNOWFLAKE:
		return proto.NewSnowflakeServiceClient(conn), nil
	case SERVICE_GEOIP:
		return proto.NewGeoIPServiceClient(conn), nil
	case SERVICE_WORDFILTER:
		return proto.NewWordFilterServiceClient(conn), nil
	case SERVICE_BGSAVE:
		return proto.NewBgSaveServiceClient(conn), nil
	}
	return nil, ERROR_SERVICE_NOT_AVAILABLE
}

func (p *service_pool) get_service(name ServiceType) (interface{}, error) {
	p.RLock()
	defer p.RUnlock()
	service := p.services[string(name)]
	if service == nil {
		return nil, ERROR_SERVICE_NOT_AVAILABLE
	}

	if len(service.clients) == 0 {
		return nil, ERROR_SERVICE_NOT_AVAILABLE
	}
	idx := int(atomic.AddUint32(&service.idx, 1))

	// add wrappers here ...
	switch name {
	case SERVICE_SNOWFLAKE:
		return proto.NewSnowflakeServiceClient(service.clients[idx%len(service.clients)].conn), nil
	case SERVICE_GEOIP:
		return proto.NewGeoIPServiceClient(service.clients[idx%len(service.clients)].conn), nil
	case SERVICE_WORDFILTER:
		return proto.NewWordFilterServiceClient(service.clients[idx%len(service.clients)].conn), nil
	case SERVICE_BGSAVE:
		return proto.NewBgSaveServiceClient(service.clients[idx%len(service.clients)].conn), nil
	}
	return nil, ERROR_SERVICE_NOT_AVAILABLE
}

// choose a service randomly
func GetService(name ServiceType) (interface{}, error) {
	return _default_pool.get_service(name)
}

// get a specific service instance with given service_name and id
func GetServiceWithId(service_name, id string) (interface{}, error) {
	return _default_pool.get_service_with_id(service_name, id)
}
