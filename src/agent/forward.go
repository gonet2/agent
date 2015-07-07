package main

import (
	"errors"
	"os"
	"registry"
	"sync"

	log "github.com/GameGophers/libs/nsq-logger"
	"github.com/GameGophers/libs/services"
	. "github.com/GameGophers/libs/services/proto"
	"golang.org/x/net/context"
)

import (
	. "types"
)

var (
	ERROR_NOT_AUTHORIZED         = errors.New("User not authorized")
	ERROR_CANNOT_FIND_GAMESERVER = errors.New("Cannot find Game Server")
)

/*
 * 长连接 双向流 RPC 到game servers, 转发agent <==> game 消息
 */

var (
	_default_forward forwarder
)

type forwarder struct {
	gs map[string]GameService_PacketClient
	sync.Mutex
}

func init() {
	_default_forward = forwarder{}
	//TODO connect all game server and forward packet
	clients, err := services.GetAllService(services.SERVICE_GAME)
	if err != nil {
		log.Critical(err)
		os.Exit(-1)
	}
	for k, cli := range clients {
		service, _ := cli.(GameServiceClient)
		stream, err := service.Packet(context.Background())
		if err != nil {
			log.Critical(err)
			os.Exit(-1)
		}
		_default_forward.gs[k] = stream

	}
	for k, c := range _default_forward.gs {
		go _default_forward.recv(k, c)
	}
}

func (f *forwarder) get(id string) GameService_PacketClient {
	f.Lock()
	defer f.Unlock()
	return f.gs[string(services.SERVICE_GAME)+"/"+id]

}

func (f *forwarder) recv(key string, cli GameService_PacketClient) {
	for {
		in, err := cli.Recv()
		if err != nil {
			log.Critical(err)
			return
		}
		registry.Deliver(in.Uid, in.Content)
	}
}

// forwarding messages to game server
func forward(sess *Session, p []byte) error {
	pkt := &Game_Packet{
		Uid:     sess.UserId,
		Content: p,
	}

	if sess.Flag&SESS_AUTHORIZED != 0 {
		// get connection
		c := _default_forward.get(sess.GSID)
		if c == nil {
			log.Criticalf("gsid %v cannot find it's game service:", sess.GSID)
			return ERROR_CANNOT_FIND_GAMESERVER
		}

		// send the packet
		if err := c.Send(pkt); err != nil {
			log.Critical(err)
			return err
		}
		return nil
	}
	return ERROR_NOT_AUTHORIZED
}
