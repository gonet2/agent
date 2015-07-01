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

/*
 * 长连接 双向流 RPC 到game servers, 转发agent <==> game 消息
 */

// message forward
func forward(sess *Session, p []byte) ([]byte, error) {
	if sess.Flag&SESS_AUTHORIZED != 0 {
		// TODO: forward packet to a game server
		return nil, nil
	}
	return nil, ERROR_NOT_AUTHORIZED
}
