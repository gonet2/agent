package ipc

import (
	"bytes"
	log "github.com/GameGophers/libs/nsq-logger"
	"gopkg.in/vmihailenco/msgpack.v2"
	"net/http"
	"os"
	"time"
)

import (
	. "types"
)

const (
	ENV_NSQD             = "NSQD_HOST"
	DEFAULT_NSQD_ADDRESS = "http://127.0.0.1:4151"
	MIME                 = "application/octet-stream"
)

var (
	_multicast_address string
	_broadcast_address string
	_unicast_address   string
)

func init() {
	// get nsqd publish address
	_unicast_address = DEFAULT_NSQD_ADDRESS + "/pub?topic=UNICAST"
	_multicast_address = DEFAULT_NSQD_ADDRESS + "/pub?topic=MULTICAST"
	_broadcast_address = DEFAULT_NSQD_ADDRESS + "/pub?topic=BROADCAST"
	if env := os.Getenv(ENV_NSQD); env != "" {
		_unicast_address = env + "/pub?topic=UNICAST"
		_multicast_address = env + "/pub?topic=MULTICAST"
		_broadcast_address = env + "/pub?topic=BROADCAST"
	}
}

// 单播一条消息到目标用户
func Unicast(src_id, dest_id int32, service int16, object interface{}, extra []byte) (ret bool) {
	// 打包为IPCObject
	req := IPCObject{SrcId: src_id,
		DestId:  dest_id,
		Service: service,
		Object:  object,
		Extra:   extra,
		Time:    time.Now().Unix()}

	// 序列化
	pack, err := msgpack.Marshal(req)
	if err != nil {
		log.Error(err)
		return false
	}

	resp, err := http.Post(_unicast_address, MIME, bytes.NewReader(pack))
	if err != nil {
		log.Critical(err)
		return false
	}
	defer resp.Body.Close()
	return true
}

// 组播
// 发消息到一组<<给定的>>目标ID
func Multicast(src_id int32, ids []int32, service int16, object interface{}, extra []byte) (ret bool) {
	// 打包为IPCObject
	req := IPCObject{SrcId: src_id,
		Service: service,
		AuxIds:  ids,
		Object:  object,
		Extra:   extra,
		Time:    time.Now().Unix()}

	// 序列化
	pack, err := msgpack.Marshal(req)
	if err != nil {
		log.Error(err)
		return false
	}

	resp, err := http.Post(_multicast_address, MIME, bytes.NewReader(pack))
	if err != nil {
		log.Critical(err)
		return false
	}
	defer resp.Body.Close()
	return true
}

// 全服广播一条动态消息
func Broadcast(src_id int32, service int16, object interface{}, extra []byte) (ret bool) {
	// 打包为IPCObject
	req := IPCObject{SrcId: src_id,
		Service: service,
		Object:  object,
		Extra:   extra,
		Time:    time.Now().Unix()}

	// 序列化
	pack, err := msgpack.Marshal(req)
	if err != nil {
		log.Error(err)
		return false
	}

	resp, err := http.Post(_broadcast_address, MIME, bytes.NewReader(pack))
	if err != nil {
		log.Critical(err)
		return false
	}
	defer resp.Body.Close()
	return true
}
