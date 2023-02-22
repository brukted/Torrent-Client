package models

import (
	"encoding/binary"
	"errors"
)

type RequestMessage struct {
	PieceIndex int
	Begin      int
	Length     int
}

func (req RequestMessage) ToBytes() []byte {
	bytes := make([]byte, 0)

	bytes = binary.BigEndian.AppendUint32(bytes, uint32(req.PieceIndex))
	bytes = binary.BigEndian.AppendUint32(bytes, uint32(req.Begin))
	bytes = binary.BigEndian.AppendUint32(bytes, uint32(req.Length))
	return bytes
}

func ReadRequestMessage(payload []byte) (RequestMessage, error) {
	if len(payload) < 12 {
		return RequestMessage{}, errors.New("invalid payload length request message")
	}
	pieceIndex := binary.BigEndian.Uint32(payload[0:4])
	begin := binary.BigEndian.Uint32(payload[4:8])
	length := binary.BigEndian.Uint32(payload[8:12])
	return RequestMessage{
		PieceIndex: int(pieceIndex),
		Begin:      int(begin),
		Length:     int(length),
	}, nil
}
