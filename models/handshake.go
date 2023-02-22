package models

import "io"

type HandShake struct {
	HeaderText string
	InfoHash   [20]byte
	PeerId     [20]byte
}

func New(infoHash, peerId [20]byte) HandShake {
	return HandShake{
		HeaderText: "BitTorrent protocol",
		InfoHash:   infoHash,
		PeerId:     peerId,
	}
}

func (handShake *HandShake) ToBytes() []byte {
	buf := make([]byte, len(handShake.HeaderText)+49)
	buf[0] = byte(len(handShake.HeaderText))
	curr := 1
	curr += copy(buf[curr:], []byte(handShake.HeaderText))
	curr += copy(buf[curr:], make([]byte, 8)) // 8 reserved bytes
	curr += copy(buf[curr:], handShake.InfoHash[:])
	curr += copy(buf[curr:], handShake.PeerId[:])
	return buf

	// bytes := make([]byte, 0)
	// bytes = append(bytes, byte(len(handShake.HeaderText)))

	// bytes = append(bytes, []byte(handShake.HeaderText)...)
	// bytes = append(bytes, make([]byte, 8)...)
	// bytes = append(bytes, handShake.InfoHash[:]...)
	// bytes = append(bytes, handShake.PeerId[:]...)

	// return bytes
}

func ReadHandShake(reader io.Reader) (HandShake, error) {
	var headerText [20]byte
	var reserved [8]byte
	var infoHash [20]byte
	var peerId [20]byte

	_, err := reader.Read(headerText[:])
	if err != nil {
		return HandShake{}, err
	}
	_, err = reader.Read(reserved[:])
	if err != nil {
		return HandShake{}, err
	}
	_, err = reader.Read(infoHash[:])
	if err != nil {
		return HandShake{}, err
	}
	_, err = reader.Read(peerId[:])
	if err != nil {
		return HandShake{}, err
	}

	return HandShake{
		HeaderText: string(headerText[:]),
		InfoHash:   infoHash,
		PeerId:     peerId,
	}, nil
}
