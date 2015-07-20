package main

import (
	"errors"

	log "github.com/gonet2/libs/nsq-logger"
	. "github.com/gonet2/libs/services/proto"
)

import (
	. "types"
)

var (
	ERROR_STREAM_NOT_OPEN = errors.New("stream not opened yet")
)

// forward messages to game server
func forward(sess *Session, p []byte) error {
	frame := &Game_Frame{
		Type:    Game_Message,
		Message: p,
	}

	// check stream
	if sess.Stream == nil {
		return ERROR_STREAM_NOT_OPEN
	}

	// forward the frame to game
	if err := sess.Stream.Send(frame); err != nil {
		log.Error(err)
		return err
	}
	return nil
}
