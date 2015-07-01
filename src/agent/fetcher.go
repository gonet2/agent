package main

import (
	"os"
	. "proto"
	"registry"

	"github.com/GameGophers/libs/services"
	"golang.org/x/net/context"
)

/*
 * 长连接 server streaming RPC 到game servers, 获得异步消息返回客户端, RPC  game server Notify
 */

// message fetcher
func fetcher_task() {
	//TODO fetch all game service;
	//connect to game service
	cli, err := services.GetService(services.SERVICE_GAME)
	if err != nil {
		log.Critical(err)
		os.Exit(-1)
	}
	service, _ := cli.(proto.GameServiceClient)
	stream, err := service.Notify(context.Background())
	if err != nil {
		log.Critical(err)
		os.Exit(-1)
	}
	for {
		p, err := stream.Recv()
		if err != nil {
			log.Critical("recv gs err", err)
			continue
		}
		registry.Deliver(p.Uid, p.Content)
	}
}
