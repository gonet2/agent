package client_handler

import (
	"crypto/rc4"
	"fmt"
	"math/big"
	"numbers"

	log "github.com/GameGophers/libs/nsq-logger"
	"github.com/GameGophers/libs/services"
)

import (
	"misc/crypto/dh"
	"misc/packet"
	. "types"
)

func checkErr(err error) {
	if err != nil {
		log.Error(err)
		panic("error occured in protocol module")
	}
}
func _client_error(sess *Session, code int) []byte {
	sess.Flag |= SESS_KICKED_OUT
	ret := command_result_info{F_code: int32(code)}
	return packet.Pack(Code["client_error_ack"], ret, nil)
}

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
	return nil
}

func P_game_servers_req(sess *Session, reader *packet.Packet) []byte {
	clients, err := services.GetAllService(services.SERVICE_GAME)
	if err != nil {
		log.Critical(err)
		return nil
	}
	table := "config@server_config"
	cnt := numbers.Count(table)
	servers := servers_info{
		F_lists: make([]server_info, 0, cnt),
	}
	for i := int32(1); i <= cnt; i++ {
		alias := numbers.GetString(table, i, "STR_Alias")
		id := string(services.SERVICE_GAME) + "/" + alias
		if _, ok := clients[id]; ok {
			name := numbers.GetString(table, i, "STR_Name")
			servers.F_lists = append(servers.F_lists, server_info{
				F_id:     i,
				F_alias:  alias,
				F_name:   name,
				F_status: GAME_SRV_OK,
			})
		}
	}

	return packet.Pack(Code["game_servers_ack"], servers, nil)
}
func P_choose_server_req(sess *Session, reader *packet.Packet) []byte {
	tbl, _ := PKT_server_alias(reader)
	alias := tbl.F_alias
	if alias == "" {
		return _client_error(sess, 0)
	}
	ok := false
	table := "config@server_config"
	cnt := numbers.Count(table)
	for i := int32(1); i <= cnt; i++ {
		if alias == numbers.GetString(table, i, "STR_Alias") {
			ok = true
			break
		}
	}
	if !ok {
		log.Critical("game server not found")
		return _client_error(sess, 0)
	}
	sess.GSID = alias
	return nil
}
