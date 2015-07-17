package client_handler

import (
	"testing"
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
	"time"
)

var (
	seqid        = uint32(0)
	encoder      *rc4.Cipher
	decoder      *rc4.Cipher
	KEY_EXCHANGE = false
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

	//user_login_req
	p3 := user_login_info{
		F_login_way:          0,
		F_open_udid:          "udid",
		F_client_certificate: "qwertyuiopasdfgh",
		F_client_version:     1,
		F_user_lang:          "en",
		F_app_id:             "com.yrhd.lovegame",
		F_os_version:         "android4.4",
		F_device_name:        "simulate",
		F_device_id:          "device_id",
		F_device_id_type:     1,
		F_login_ip:           "127.0.0.1",
	}
	send_proto(conn, Code["user_login_req"], p3)

	//heart_beat_req
	send_proto(conn, Code["heart_beat_req"], nil)

	//proto_ping_req
	p1 := auto_id{
		F_id: rand.Int31(),
	}
	send_proto(conn, Code["proto_ping_req"], p1)

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

