package client_handler

import (
	"crypto/rc4"
	"fmt"
	"io"
	"math/big"

	"golang.org/x/net/context"

	log "github.com/gonet2/libs/nsq-logger"
)

import (
	"misc/crypto/dh"
	"misc/packet"
	. "types"

	sp "github.com/gonet2/libs/services"
	spp "github.com/gonet2/libs/services/proto"
)

// 心跳包
func P_heart_beat_req(sess *Session, reader *packet.Packet) []byte {
	tbl, _ := PKT_auto_id(reader)
	return packet.Pack(Code["heart_beat_ack"], tbl, nil)
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
	// TODO: 登陆鉴权
	sess.UserId = 1

	// TODO: 选择登陆服务器
	sess.GSID = DEFAULT_GSID

	// 选服
	cli, err := sp.GetServiceWithId(sp.SERVICE_GAME, sess.GSID)
	if err != nil {
		log.Critical(err)
		return nil
	}

	// type assertion
	game_cli, ok := cli.(spp.GameServiceClient)
	if !ok {
		log.Critical("cannot do type assertion on: %v", sess.GSID)
		return nil
	}

	// 开启到游戏服的流
	// TODO: 处理context，设定超时
	stream, err := game_cli.Stream(context.Background())
	if err != nil {
		log.Critical(err)
		return nil
	}
	sess.Stream = stream

	// 在game注册
	// TODO: 新用户的创建由game处理
	sess.Stream.Send(&spp.Game_Frame{Type: spp.Game_Register, UserId: sess.UserId})

	// 读取GAME返回消息
	fetcher_task := func(sess *Session) {
		for {
			in, err := sess.Stream.Recv()
			if err == io.EOF { // 流关闭
				log.Trace(err)
				return
			}
			if err != nil {
				log.Error(err)
				return
			}
			sess.MQ <- *in
		}
	}
	go fetcher_task(sess)
	return packet.Pack(Code["user_login_ack"], user_snapshot{F_uid: sess.UserId}, nil)
}

func checkErr(err error) {
	if err != nil {
		log.Error(err)
		panic("error occured in protocol module")
	}
}
