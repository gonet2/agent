package main

import (
	"errors"
	"io"
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
	ERROR_NOT_AUTHORIZED = errors.New("User not authorized")
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
		if err == io.EOF {
			log.Infof("stream recv EOF err : %v", err)
			return
		}
		if err != nil {
			log.Critical("Failed to receive a note : %v", err)
			continue
		}
		registry.Deliver(in.Uid, in.Content)
	}
}

// message forward
func forward(sess *Session, p []byte) ([]byte, error) {
	if sess.Flag&SESS_AUTHORIZED != 0 {
		// TODO: forward packet to a game server
		pkt := &Game_Packet{
			Uid:     sess.UserId,
			Content: p,
		}
		if err := _default_forward.get(sess.GSID).Send(pkt); err != nil {
			log.Criticalf("Failed to send pkt %v", err)
		}
	}
	return nil, ERROR_NOT_AUTHORIZED
}
