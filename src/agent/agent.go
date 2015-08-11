package main

import (
	spp "github.com/gonet2/libs/services/proto"
	"time"
)

import (
	. "types"
	"utils"
)

// agent of user
func agent(sess *Session, in chan []byte, out *Buffer, sess_die chan struct{}) {
	defer wg.Done()
	defer utils.PrintPanicStack()

	// init session
	sess.MQ = make(chan spp.Game_Frame, DEFAULT_MQ_SIZE)
	sess.ConnectTime = time.Now()
	sess.LastPacketTime = time.Now()
	// minute timer
	min_timer := time.After(time.Minute)

	// cleanup work
	defer func() {
		close(sess_die)
		if sess.Stream != nil {
			sess.Stream.CloseSend()
		}
	}()

	// >> the main message loop <<
	for {
		select {
		case msg, ok := <-in: // packet from network
			if !ok {
				return
			}

			sess.PacketCount++
			sess.PacketTime = time.Now()

			if result := proxy_user_request(sess, msg); result != nil {
				out.send(sess, result)
			}
			sess.LastPacketTime = sess.PacketTime
		case frame := <-sess.MQ:
			switch frame.Type {
			case spp.Game_Message:
				out.send(sess, frame.Message)
			case spp.Game_Kick:
				sess.Flag |= SESS_KICKED_OUT
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
