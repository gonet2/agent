package packet

import (
	log "github.com/GameGophers/libs/nsq-logger"
	"github.com/golang/protobuf/proto"
)

type FastPack interface {
	Pack(w *Packet)
}

// export struct fields with packet writer.
func Pack(tos int16, tbl interface{}, writer *Packet) []byte {
	// create writer if not specified
	if writer == nil {
		writer = Writer()
	}

	// write protocol number
	if tos != -1 {
		writer.WriteU16(uint16(tos))
	}

	// is the table nil?
	if tbl == nil {
		return writer.Data()
	}
	p, err := proto.Marshal(tbl.(proto.Message))
	if err == nil {
		log.Critical(err)
		return nil
	}
	writer.WriteRawBytes(p)

	// return byte array
	return writer.Data()
}
