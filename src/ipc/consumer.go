package ipc

import (
	"fmt"
	log "github.com/GameGophers/nsq-logger"
	nsq "github.com/bitly/go-nsq"
	"github.com/coreos/go-etcd/etcd"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"gopkg.in/vmihailenco/msgpack.v2"
	"math/rand"
	"os"
	"strings"
	"sync"
)

import (
	"registry"
	services "services/proto"
	. "types"
)

const (
	DEFAULT_NSQLOOKUPD = "127.0.0.1:4160"
	ENV_NSQLOOKUPD     = "NSQLOOKUPD_HOST"
	DEFAULT_ETCD       = "127.0.0.1:2379"
	NSQ_IN_FLIGHT      = 128
)

const (
	UNICAST   = "unicast"
	MULTICAST = "multicast"
	BROADCAST = "broadcast"
)

const (
	SNOWFLAKE_SERVICE_NAME = "/backends/snowflake"
)

func init() {
	ic := IPCConsumer{}
	ic.init()
}

type IPCConsumer struct {
	client_pool sync.Pool
}

func (ic *IPCConsumer) init() {
	// etcd client
	machines := []string{DEFAULT_ETCD}
	if env := os.Getenv("ETCD_HOST"); env != "" {
		machines = strings.Split(env, ";")
	}
	ic.client_pool.New = func() interface{} {
		return etcd.NewClient(machines)
	}

	// rank change subscriber
	channel := ic.create_ephermal_channel()
	ic.init_subscriber(UNICAST, channel)
	ic.init_subscriber(MULTICAST, channel)
	ic.init_subscriber(BROADCAST, channel)
}

func (ic *IPCConsumer) create_ephermal_channel() string {
	client := ic.client_pool.Get().(*etcd.Client)
	defer func() {
		ic.client_pool.Put(client)
	}()
	resp, err := client.Get(SNOWFLAKE_SERVICE_NAME, false, true)
	if err != nil {
		log.Critical(err)
		os.Exit(-1)
	}

	// random choose a service
	if len(resp.Node.Nodes) == 0 {
		log.Critical("snowflake service not started yet?")
		os.Exit(-1)
	}

	// dial grpc
	conn, err := grpc.Dial(resp.Node.Nodes[rand.Intn(len(resp.Node.Nodes))].Value)
	if err != nil {
		log.Critical("did not connect: %v", err)
		os.Exit(-1)
	}

	// save client
	c := services.NewSnowflakeServiceClient(conn)
	// create consumer from lookupds
	r, err := c.GetUUID(context.Background(), &services.Snowflake_NullRequest{})
	if err != nil {
		log.Critical(err)
		os.Exit(-1)
	}
	defer conn.Close()

	return fmt.Sprintf("IPC%v#ephemeral", r.Uuid)
}

func (ic *IPCConsumer) init_subscriber(topic, channel string) {
	cfg := nsq.NewConfig()
	cfg.MaxInFlight = NSQ_IN_FLIGHT
	consumer, err := nsq.NewConsumer(topic, channel, cfg)
	if err != nil {
		log.Critical(err)
		os.Exit(-1)
	}

	// message process
	consumer.AddHandler(nsq.HandlerFunc(func(msg *nsq.Message) error {
		switch topic {
		case UNICAST:
			ic.unicast(msg)
		case MULTICAST:
			ic.multicast(msg)
		case BROADCAST:
			ic.broadcast(msg)
		}
		return nil
	}))

	// read environtment variable
	addresses := []string{DEFAULT_NSQLOOKUPD}
	if env := os.Getenv(ENV_NSQLOOKUPD); env != "" {
		addresses = strings.Split(env, ";")
	}

	// connect to nsqlookupd
	log.Trace("connect to nsqlookupds ip:", addresses)
	if err := consumer.ConnectToNSQLookupds(addresses); err != nil {
		log.Critical(err)
		return
	}
	log.Info("nsqlookupd connected for unicast")
}

func (ic *IPCConsumer) unicast(msg *nsq.Message) {
	// 解包
	obj := IPCObject{}
	err := msgpack.Unmarshal(msg.Body, &obj)
	if err != nil {
		log.Error(err)
		return
	}

	// 投递
	registry.Deliver(obj.DestId, obj)
}

func (ic *IPCConsumer) multicast(msg *nsq.Message) {
	// 解包
	obj := IPCObject{}
	err := msgpack.Unmarshal(msg.Body, &obj)
	if err != nil {
		log.Error(err)
		return
	}

	// 循环投递
	for _, id := range obj.AuxIds {
		registry.Deliver(id, obj)
	}
}

func (ic *IPCConsumer) broadcast(msg *nsq.Message) {
	// 解包
	obj := IPCObject{}
	err := msgpack.Unmarshal(msg.Body, &obj)
	if err != nil {
		log.Error(err)
		return
	}

	// 循环投递
	users := registry.ListAll()
	for _, v := range users {
		registry.Deliver(v, obj)
	}
}
