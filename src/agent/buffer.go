package main

import (
	"encoding/binary"
	"net"
	"time"

	log "github.com/GameGophers/libs/nsq-logger"
)

import (
	"misc/packet"
	. "types"
	"utils"
)

type Buffer struct {
	ctrl    chan bool    // receive exit signal
	pending chan []byte  // pending packets
	conn    *net.TCPConn // connection
	cache   []byte       // for combined syscall write
}

var (
	// for padding packet, random content
	_padding [PADDING_SIZE]byte
)

func init() {
	go func() { // padding content update procedure
		for {
			for k := range _padding {
				_padding[k] = byte(<-utils.LCG)
			}
			log.Info("Padding Updated:", _padding)
			<-time.After(PADDING_UPDATE_PERIOD * time.Second)
		}
	}()
}

// packet sending procedure
func (buf *Buffer) send(sess *Session, data []byte) {
	// in case of empty packet
	if data == nil {
		return
	}

	// padding
	// if the size of the data to return is tiny, pad with some random numbers
	if len(data) < PADDING_LIMIT {
		data = append(data, _padding[:]...)
	}

	// encryption
	if sess.Flag&SESS_ENCRYPT != 0 { // encryption is enabled
		sess.Encoder.XORKeyStream(data, data)
	} else if sess.Flag&SESS_KEYEXCG != 0 { // key is exchanged, encryption is not yet enabled
		sess.Flag &^= SESS_KEYEXCG
		sess.Flag |= SESS_ENCRYPT
	}

	// queue the data for sending
	buf.pending <- data
	return
}

// packet sending goroutine
func (buf *Buffer) start() {
	defer utils.PrintPanicStack()
	for {
		select {
		case data := <-buf.pending:
			buf.raw_send(data)
		case <-buf.ctrl: // receive session end signal
			close(buf.pending)
			// close the connection
			buf.conn.Close()
			return
		}
	}
}

// raw packet encapsulation and put it online
func (buf *Buffer) raw_send(data []byte) bool {
	// combine output
	sz := len(data)
	binary.BigEndian.PutUint16(buf.cache, uint16(sz))
	copy(buf.cache[2:], data)

	// write data
	n, err := buf.conn.Write(buf.cache[:sz+2])
	if err != nil {
		log.Warningf("Error send reply data, bytes: %v reason: %v", n, err)
		return false
	}

	return true
}

// set send buffer size
func (buf *Buffer) set_write_buffer(bytes int) {
	buf.conn.SetWriteBuffer(bytes)
}

// create a associated write buffer for a session
func new_buffer(conn *net.TCPConn, ctrl chan bool) *Buffer {
	buf := Buffer{conn: conn}
	buf.pending = make(chan []byte)
	buf.ctrl = ctrl
	buf.cache = make([]byte, packet.PACKET_LIMIT+2)
	return &buf
}
