package main

import (
	"io"
	"os"

	"golang.org/x/net/context"
)
import (
	"registry"

	log "github.com/GameGophers/libs/nsq-logger"
	"github.com/GameGophers/libs/services"
	"github.com/GameGophers/libs/services/proto"
)

/*
 * 长连接 server streaming RPC 到game servers, 获得异步消息返回客户端, RPC  game server Notify
 */

// message fetcher
func fetcher_task() {
	//connect to game service
	clients, err := services.GetAllService(services.SERVICE_GAME)
	if err != nil {
		log.Critical(err)
		os.Exit(-1)
	}
	for _, cli := range clients {
		service, _ := cli.(proto.GameServiceClient)
		stream, err := service.Notify(context.Background(), &proto.Game_Nil{})
		if err != nil {
			log.Critical(err)
			os.Exit(-1)
		}
		go func() {
			for {
				p, err := stream.Recv()
				if err == io.EOF {
					log.Infof("stream recv EOF err :%v", err)
					return
				}
				if err != nil {
					log.Critical("stream recv gs err", err)
					continue
				}
				registry.Deliver(p.Uid, p.Content)
			}
		}()
	}
}
