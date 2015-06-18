package client_handler

import (
	"logic"
	"math/big"

	"golang.org/x/net/context"
)

import (
	"misc/crypto/dh"
	"misc/crypto/pike"
	"misc/packet"
	. "types"

	log "github.com/GameGophers/libs/nsq-logger"
	"github.com/GameGophers/libs/services"
	proto "github.com/GameGophers/libs/services/proto"
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
	tbl, _ := PKT_user_login_info(reader)
	cli, err := services.GetService(services.SERVICE_LOGIN)
	if err != nil {
		log.Critical(err)
		return packet.Pack(Code["command_result_info"], command_result_info{F_code: 1, F_msg: "login service err"}, nil)
	}
	service, _ := cli.(proto.LoginServiceClient)
	user_login := &proto.User_LoginInfo{
		Uuid:      tbl.F_open_udid,
		Host:      "agent1",
		LoginType: tbl.F_login_way,
		Username:  tbl.F_open_udid,
		Passwd:    tbl.F_client_certificate,
	}
	r, err := service.Login(context.Background(), user_login)
	if err != nil {
		log.Critical(err)
		return packet.Pack(Code["command_result_info"], command_result_info{F_code: 0, F_msg: "login faild"}, nil)
	}
	sess.UserId = r.Uid
	user := &User{}
	if r.NewUser == true {
		if err := logic.UserInit(r.Uid, sess); err != nil {
			log.Critical("init user error %v", err)
			return nil
		}
	} else {
		if err := logic.UserLoad(r.Uid, sess); err != nil {
			log.Critical("init user error %v", err)
			return nil
		}
	}

	return packet.Pack(Code["user_snapshot"], user_snapshot{F_uid: user.Id, F_name: user.Name, F_level: int32(user.Level)}, nil)
}

func checkErr(err error) {
	if err != nil {
		log.Error(err)
		panic("error occured in protocol module")
	}
}
