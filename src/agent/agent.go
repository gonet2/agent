package main

import (
	"io"
	"misc/packet"
	"time"

	log "github.com/GameGophers/libs/nsq-logger"
	"github.com/GameGophers/libs/services"
	"github.com/GameGophers/libs/services/proto"
	"golang.org/x/net/context"
)

import (
	. "types"
	"utils"
)

// agent of user
func agent(sess *Session, in chan []byte, out *Buffer, sess_die chan bool) {
	defer wg.Done()
	defer utils.PrintPanicStack()

	// init session
	sess.MQ = make(chan IPCObject, DEFAULT_MQ_SIZE)
	sess.ConnectTime = time.Now()
	sess.LastPacketTime = time.Now()
	// minute timer
	min_timer := time.After(time.Minute)

	// cleanup work
	defer func() {
		close_work(sess)
		close(sess_die)
	}()

	//proxy the request to game service.
	cli, err := services.GetService(services.SERVICE_GAME)
	if err != nil {
		log.Critical(err)
		return
	}
	service, _ := cli.(proto.GameServiceClient)
	stream, err := service.Packet(context.Background())
	if err != nil {
		log.Critical(err)
		return
	}
	defer stream.CloseSend()

	// >> the main message loop <<
	for {
		select {
		case msg, ok := <-in: // packet from network
			if !ok {
				return
			}

			sess.PacketCount++
			sess.PacketTime = time.Now()

			reader := packet.Reader(msg)
			reader.ReadU32()
			c, _ := reader.ReadS16()
			switch {
			case c <= 1000:
				if result := proxy_user_request(sess, msg); result != nil {
					out.send(sess, result)
				}
			default:
				p := &proto.Game_Packet{
					Uid:     sess.UserId,
					Content: reader.Data(),
				}
				go func() {
					recv, err := stream.Recv()
					if err == io.EOF {
						log.Criticalf("EOF recv %v", err)
						return
					}
					if err != nil {
						log.Criticalf("Failed to recv pkt %v", err)

					}
					out.send(sess, recv.Content)
				}()
				if err = stream.Send(p); err != nil {
					log.Criticalf("Failed to send pkt %v", err)
				}
			}

			sess.LastPacketTime = sess.PacketTime
		case msg := <-sess.MQ: // message from server internal IPC
			if result := proxy_ipc_request(sess, msg); result != nil {
				out.send(sess, result)
			}
		case <-min_timer: // minutes timer
			timer_work(sess, out)
			min_timer = time.After(time.Minute)
		case <-die: // server is shuting down...
			sess.Flag |= SESS_KICKED_OUT
		}

		// see if the player should be kicked out.
		if sess.Flag&SESS_KICKED_OUT != 0 {
			return
		}
	}
}
