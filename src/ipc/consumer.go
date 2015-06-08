package ipc

import (
	"fmt"
	log "github.com/GameGophers/nsq-logger"
	nsq "github.com/bitly/go-nsq"
	"golang.org/x/net/context"
	"gopkg.in/vmihailenco/msgpack.v2"
	"os"
	"strings"
)

import (
	"registry"
	"services"
	"services/proto"
	. "types"
)

const (
	DEFAULT_NSQLOOKUPD = "http://127.0.0.1:4161"
	ENV_NSQLOOKUPD     = "NSQLOOKUPD_HOST"
	NSQ_IN_FLIGHT      = 128
)

const (
	UNICAST   = "unicast"
	MULTICAST = "multicast"
	BROADCAST = "broadcast"
)

func init() {
	ic := IPCConsumer{}
	ic.init()
}

type IPCConsumer struct{}

func (ic *IPCConsumer) init() {
	// rank change subscriber
	channel := ic.create_ephermal_channel()
	ic.init_subscriber(UNICAST, channel)
	ic.init_subscriber(MULTICAST, channel)
	ic.init_subscriber(BROADCAST, channel)
}

func (ic *IPCConsumer) create_ephermal_channel() string {
	c, err := services.GetService(services.SERVICE_SNOWFLAKE)
	if err != nil {
		log.Critical(err)
		os.Exit(-1)
	}
	service, ok := c.(proto.SnowflakeServiceClient)
	if !ok {
		log.Critical("internal error:", c)
		os.Exit(-1)
	}

	r, err := service.GetUUID(context.Background(), &proto.Snowflake_NullRequest{})
	if err != nil {
		log.Critical(err)
		os.Exit(-1)
	}
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
	log.Info("nsqlookupd connected")
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
