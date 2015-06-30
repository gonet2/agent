package main

import (
	"errors"
)

import (
	. "types"
)

var (
	ERROR_NOT_AUTHORIZED = errors.New("User not authorized")
)

func forward(sess *Session, p []byte) ([]byte, error) {
	if sess.Flag&SESS_AUTHORIZED != 0 {
		// TODO: forward packet to a game server
		return nil, nil
	}
	return nil, ERROR_NOT_AUTHORIZED
}
