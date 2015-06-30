package client_handler

import (
	log "github.com/GameGophers/nsq-logger"
	"math/big"
)

import (
	"misc/crypto/dh"
	"misc/crypto/pike"
	"misc/packet"
	. "types"
)

// 心跳包
func P_heart_beat_req(sess *Session, reader *packet.Packet) []byte {
	return packet.Pack(Code["heart_beat_ack"], nil, nil)
}

// 密钥交换
func P_get_pike_seed_req(sess *Session, reader *packet.Packet) []byte {
	tbl, _ := PKT_pike_seed_info(reader)
	// KEY1
	X1, E1 := dh.DHExchange()
	KEY1 := dh.DHKey(X1, big.NewInt(int64(tbl.F_client_send_seed)))

	// KEY2
	X2, E2 := dh.DHExchange()
	KEY2 := dh.DHKey(X2, big.NewInt(int64(tbl.F_client_receive_seed)))

	ret := pike_seed_info{int32(E1.Int64()), int32(E2.Int64())}
	// 服务器加密种子是客户端解密种子
	sess.Encoder = pike.NewCtx(uint32(KEY2.Int64()))
	sess.Decoder = pike.NewCtx(uint32(KEY1.Int64()))
	sess.Flag |= SESS_KEYEXCG
	return packet.Pack(Code["get_pike_seed_ack"], ret, nil)
}

// 玩家登陆过程
func P_user_login_req(sess *Session, reader *packet.Packet) []byte {
	return nil
}

func checkErr(err error) {
	if err != nil {
		log.Error(err)
		panic("error occured in protocol module")
	}
}
