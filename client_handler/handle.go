package client_handler

import (
	"crypto/rc4"
	"fmt"
	"io"
	"math/big"

	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"

	"agent/misc/crypto/dh"
	"agent/misc/packet"
	"agent/services"

	log "github.com/Sirupsen/logrus"

	pb "agent/pb"

	. "agent/types"
)

// 心跳包
func P_heart_beat_req(sess *Session, reader *packet.Packet) []byte {
	tbl, _ := PKT_auto_id(reader)
	return packet.Pack(Code["heart_beat_ack"], tbl, nil)
}

// 密钥交换
// 加密建立方式: DH+RC4
// 注意:完整的加密过程包括 RSA+DH+RC4
// 1. RSA用于鉴定服务器的真伪(这步省略)
// 2. DH用于在不安全的信道上协商安全的KEY
// 3. RC4用于流加密
func P_get_seed_req(sess *Session, reader *packet.Packet) []byte {
	tbl, _ := PKT_seed_info(reader)
	// KEY1
	X1, E1 := dh.DHExchange()
	KEY1 := dh.DHKey(X1, big.NewInt(int64(tbl.F_client_send_seed)))

	// KEY2
	X2, E2 := dh.DHExchange()
	KEY2 := dh.DHKey(X2, big.NewInt(int64(tbl.F_client_receive_seed)))

	ret := S_seed_info{int32(E1.Int64()), int32(E2.Int64())}
	// 服务器加密种子是客户端解密种子
	encoder, err := rc4.NewCipher([]byte(fmt.Sprintf("%v%v", SALT, KEY2)))
	if err != nil {
		log.Error(err)
		return nil
	}
	decoder, err := rc4.NewCipher([]byte(fmt.Sprintf("%v%v", SALT, KEY1)))
	if err != nil {
		log.Error(err)
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
	// 简单鉴权可以在agent直接完成，通常公司都存在一个用户中心服务器用于鉴权
	sess.UserId = 1

	// TODO: 选择GAME服务器
	// 选服策略依据业务进行，比如小服可以固定选取某台，大服可以采用HASH或一致性HASH
	sess.GSID = DEFAULT_GSID

	// 连接到已选定GAME服务器
	conn := services.GetServiceWithId("game-10000", sess.GSID)
	if conn == nil {
		log.Error("cannot get game service:", sess.GSID)
		return nil
	}
	cli := pb.NewGameServiceClient(conn)

	// 开启到游戏服的流
	ctx := metadata.NewContext(context.Background(), metadata.New(map[string]string{"userid": fmt.Sprint(sess.UserId)}))
	stream, err := cli.Stream(ctx)
	if err != nil {
		log.Error(err)
		return nil
	}
	sess.Stream = stream

	// 读取GAME返回消息的goroutine
	fetcher_task := func(sess *Session) {
		for {
			in, err := sess.Stream.Recv()
			if err == io.EOF { // 流关闭
				log.Debug(err)
				return
			}
			if err != nil {
				log.Error(err)
				return
			}
			select {
			case sess.MQ <- *in:
			case <-sess.Die:
			}
		}
	}
	go fetcher_task(sess)
	return packet.Pack(Code["user_login_succeed_ack"], S_user_snapshot{F_uid: sess.UserId}, nil)
}
