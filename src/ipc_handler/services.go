package ipc_handler

import (
	. "types"
)

const (
	SVC_CHAT = 1
)

var Handlers map[int16]func(*Session, IPCObject) []byte = map[int16]func(*Session, IPCObject) []byte{
	SVC_CHAT: ipc_chat,
}
