package packet

import (
	"errors"

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

	//use protobuf marshal pack the data.
	if pb, ok := tbl.(proto.Message); ok {
		p, err := proto.Marshal(pb)
		if err == nil {
			log.Critical(err)
			return nil
		}
		writer.WriteRawBytes(p)
	} else {
		log.Critical("tbl %+v is not implement proto.Message interface", tbl)
		return nil
	}

	// return byte array
	return writer.Data()
}

//unpack the data
func Unpack(in *Packet, out interface{}) error {
	msg, ok := out.(*proto.Message)
	if !ok {
		return errors.New("unpack data must be a *proto.Message")
	}
	bin, err := in.ReadRawBytes()
	if err != nil {
		return err
	}
	err = proto.Unmarshal(bin, *msg)
	if err != nil {
		return err
	}
	return nil
}
