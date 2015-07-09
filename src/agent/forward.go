package main

import (
	"errors"
	"io"

	log "github.com/GameGophers/libs/nsq-logger"
	. "github.com/GameGophers/libs/services/proto"
)

import (
	. "types"
)

var (
	ERROR_NOT_AUTHORIZED = errors.New("User not authorized")
)

// forward messages to game server
func forward(sess *Session, p []byte) error {
	pkt := &Game_Packet{
		Ctrl:    Game_Message,
		Content: p,
	}

	if sess.Flag&SESS_AUTHORIZED != 0 {
		// send the packet
		if err := sess.Stream.Send(pkt); err != nil {
			log.Critical(err)
			return err
		}
		return nil
	}
	return ERROR_NOT_AUTHORIZED
}

// fetch messages for current session
func fetcher_task(sess *Session) {
	for {
		in, err := sess.Stream.Recv()
		// close signal
		if err == io.EOF {
			log.Trace(err)
			return
		}
		if err != nil {
			log.Critical(err)
			return
		}

		switch in.Ctrl {
		case Game_Message:
			sess.MQ <- in.Content
		}
	}
}
