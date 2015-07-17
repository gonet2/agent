package main

import (
	"crypto/rc4"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/big"
	"math/rand"
	"misc/crypto/dh"
	"misc/packet"
	"net"
	"os"
	"testing"
	"time"
)

var (
	seqid        = uint32(0)
	encoder      *rc4.Cipher
	decoder      *rc4.Cipher
	KEY_EXCHANGE = false
	SALT         = "DH"
)

const (
	DEFAULT_AGENT_HOST = "127.0.0.1:8888"
)

func Test_agent(t *testing.T) {
	host := DEFAULT_AGENT_HOST
	if env := os.Getenv("AGENT_HOST"); env != "" {
		host = env
	}
	addr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
	defer conn.Close()

	//get_seed_req
	S1, M1 := dh.DHExchange()
	S2, M2 := dh.DHExchange()
	p2 := seed_info{
		int32(M1.Int64()),
		int32(M2.Int64()),
	}
	rst := send_proto(conn, Code["get_seed_req"], p2)
	r1, _ := PKT_seed_info(rst)
	log.Printf("result: %#v", r1)

	K1 := dh.DHKey(S1, big.NewInt(int64(r1.F_client_send_seed)))
	K2 := dh.DHKey(S2, big.NewInt(int64(r1.F_client_receive_seed)))
	encoder, err = rc4.NewCipher([]byte(fmt.Sprintf("%v%v", SALT, K1)))
	if err != nil {
		log.Println(err)
		return
	}
	decoder, err = rc4.NewCipher([]byte(fmt.Sprintf("%v%v", SALT, K2)))
	if err != nil {
		log.Println(err)
		return
	}

	KEY_EXCHANGE = true

	//heart_beat_req
	p1 := auto_id{F_id: rand.Int31()}
	log.Printf("send: %#v", p1)
	rst = send_proto(conn, Code["heart_beat_req"], p1)
	r2, _ := PKT_auto_id(rst)
	log.Printf("result: %#v", r2)

}

func send_proto(conn net.Conn, p int16, info interface{}) (reader *packet.Packet) {
	seqid++
	payload := packet.Pack(p, info, nil)
	writer := packet.Writer()
	writer.WriteU16(uint16(len(payload)) + 4)

	w := packet.Writer()
	w.WriteU32(seqid)
	w.WriteRawBytes(payload)
	data := w.Data()
	if KEY_EXCHANGE {
		encoder.XORKeyStream(data, data)
	}
	writer.WriteRawBytes(data)
	conn.Write(writer.Data())
	log.Printf("send : %#v", writer.Data())
	time.Sleep(time.Second)

	//read
	header := make([]byte, 2)
	io.ReadFull(conn, header)
	size := binary.BigEndian.Uint16(header)
	log.Printf("read header: %v \n", size)
	r := make([]byte, size)
	_, err := io.ReadFull(conn, r)
	if err != nil {
		log.Println(err)
	}
	if KEY_EXCHANGE {
		decoder.XORKeyStream(r, r)
	}
	reader = packet.Reader(r)
	b, err := reader.ReadS16()
	if err != nil {
		log.Println(err)
	}
	if _, ok := RCode[b]; !ok {
		log.Println("unknown proto ", b)
	}

	return
}

var Code = map[string]int16{
	"heart_beat_req":         0,  // 心跳包..
	"heart_beat_ack":         1,  // 心跳包回复
	"user_login_req":         10, // 登陆
	"user_login_succeed_ack": 11, // 登陆成功
	"user_login_faild_ack":   12, // 登陆失败
	"client_error_ack":       13, // 客户端错误
	"get_seed_req":           30, // socket通信加密使用
	"get_seed_ack":           31, // socket通信加密使用
}

var RCode = map[int16]string{
	0:  "heart_beat_req",         // 心跳包..
	1:  "heart_beat_ack",         // 心跳包回复
	10: "user_login_req",         // 登陆
	11: "user_login_succeed_ack", // 登陆成功
	12: "user_login_faild_ack",   // 登陆失败
	13: "client_error_ack",       // 客户端错误
	30: "get_seed_req",           // socket通信加密使用
	31: "get_seed_ack",           // socket通信加密使用
}

type auto_id struct {
	F_id int32
}

func (p auto_id) Pack(w *packet.Packet) {
	w.WriteS32(p.F_id)

}

func PKT_auto_id(reader *packet.Packet) (tbl auto_id, err error) {
	tbl.F_id, err = reader.ReadS32()
	checkErr(err)

	return
}

type seed_info struct {
	F_client_send_seed    int32
	F_client_receive_seed int32
}

func (p seed_info) Pack(w *packet.Packet) {
	w.WriteS32(p.F_client_send_seed)
	w.WriteS32(p.F_client_receive_seed)

}
func PKT_seed_info(reader *packet.Packet) (tbl seed_info, err error) {
	tbl.F_client_send_seed, err = reader.ReadS32()
	checkErr(err)

	tbl.F_client_receive_seed, err = reader.ReadS32()
	checkErr(err)

	return
}
func checkErr(err error) {
	if err != nil {
		log.Println(err)
		panic("error occured in protocol module")
	}
}
