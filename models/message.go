package models

import "encoding/binary"

type MessageType byte

const (
	MsgTypeChoke         MessageType = 0
	MsgTypeUnChoke       MessageType = 1
	MsgTypeInterested    MessageType = 2
	MsgTypeNotInterested MessageType = 3
	MsgTypeHave          MessageType = 4
	MsgTypeBitField      MessageType = 5
	MsgTypeRequest       MessageType = 6
	MsgTypePiece         MessageType = 7
	MsgTypeCancel        MessageType = 8
	MsgTypeKeepAlive     MessageType = 9
)

type Message struct {
	Type    MessageType
	Payload []byte
}

func (m *Message) ToBytes() []byte {
	if m == nil {
		return make([]byte, 4)
	}

	length := uint32(len(m.Payload) + 1)
	buffer := make([]byte, length+4)
	binary.BigEndian.PutUint32(buffer[0:4], length)

	buffer[4] = byte(m.Type)
	copy(buffer[5:], m.Payload)

	return buffer
}

func (mType MessageType) String() string {
	switch mType {
	case MsgTypeChoke:
		return "Choke"
	case MsgTypeUnChoke:
		return "UnChoke"
	case MsgTypeInterested:
		return "Interested"
	case MsgTypeNotInterested:
		return "NotInterested"
	case MsgTypeHave:
		return "Have"
	case MsgTypeBitField:
		return "BitField"
	case MsgTypeRequest:
		return "Request"
	case MsgTypePiece:
		return "Piece"
	case MsgTypeCancel:
		return "Cancel"
	default:
		return "Unknown"
	}
}
