package client_handler

import (
	"crypto/rc4"
	"fmt"
	"io"
	"math/big"

	log "github.com/gonet2/libs/nsq-logger"
)

import (
	sp "github.com/gonet2/libs/services/proto"
	"misc/crypto/dh"
	"misc/packet"
	. "types"
)

// 心跳包
func P_heart_beat_req(sess *Session, reader *packet.Packet) []byte {
	return packet.Pack(Code["heart_beat_ack"], nil, nil)
}

// 密钥交换
func P_get_seed_req(sess *Session, reader *packet.Packet) []byte {
	tbl, _ := PKT_seed_info(reader)
	// KEY1
	X1, E1 := dh.DHExchange()
	KEY1 := dh.DHKey(X1, big.NewInt(int64(tbl.F_client_send_seed)))

	// KEY2
	X2, E2 := dh.DHExchange()
	KEY2 := dh.DHKey(X2, big.NewInt(int64(tbl.F_client_receive_seed)))

	ret := seed_info{int32(E1.Int64()), int32(E2.Int64())}
	// 服务器加密种子是客户端解密种子
	encoder, err := rc4.NewCipher([]byte(fmt.Sprintf("%v%v", SALT, KEY2)))
	if err != nil {
		log.Critical(err)
		return nil
	}
	decoder, err := rc4.NewCipher([]byte(fmt.Sprintf("%v%v", SALT, KEY1)))
	if err != nil {
		log.Critical(err)
		return nil
	}
	sess.Encoder = encoder
	sess.Decoder = decoder
	sess.Flag |= SESS_KEYEXCG
	return packet.Pack(Code["get_seed_ack"], ret, nil)
}

// 玩家登陆过程
func P_user_login_req(sess *Session, reader *packet.Packet) []byte {
	// fetch messages for current session
	fetcher_task := func(sess *Session) {
		defer func() {
			log.Tracef("fetcher task closed userid:%v", sess.UserId)
		}()
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
			case sp.Game_Message:
				sess.MQ <- in.Message
			case sp.Game_Kick:
				sess.Flag |= SESS_KICKED_OUT
			}
		}
	}
	go fetcher_task(sess)
	return nil
}

func checkErr(err error) {
	if err != nil {
		log.Error(err)
		panic("error occured in protocol module")
	}
}
