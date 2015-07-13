package main

import (
	"errors"
	"io"

	log "github.com/gonet2/libs/nsq-logger"
	. "github.com/gonet2/libs/services/proto"
)

import (
	. "types"
)

var (
	ERROR_NOT_AUTHORIZED = errors.New("User not authorized")
)

// forward messages to game server
func forward(sess *Session, p []byte) error {
	frame := &Game_Frame{
		Type:    Game_Message,
		Message: p,
	}

	if sess.Flag&SESS_AUTHORIZED != 0 {
		// send the packet
		if err := sess.Stream.Send(frame); err != nil {
			log.Error(err)
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
			log.Error(err)
			return
		}

		switch in.Type {
		case Game_Message:
			sess.MQ <- in.Message
		case Game_Kick:
			sess.Flag |= SESS_KICKED_OUT
		}
	}
}
