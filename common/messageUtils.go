package common

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"torrentClient/models"
)

func Read(r io.Reader) (*models.Message, error) {
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBuf)

	// keep-alive message
	if length == 0 {
		return nil, nil
	}

	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}

	message := models.Message{
		Type:    models.MessageType(messageBuf[0]),
		Payload: messageBuf[1:],
	}

	return &message, nil
}

func SendHaveMessage(peer *models.Peer, pieceIndex int) (err error) {
	_, err = SendMessageWithRetry(peer, models.Message{
		Type:    models.MsgTypeHave,
		Payload: binary.BigEndian.AppendUint32([]byte{}, uint32(pieceIndex)),
	})
	return
}

func SendUnchokeMessage(peer *models.Peer) (err error) {
	_, err = SendMessageWithRetry(peer, models.Message{
		Type:    models.MsgTypeUnChoke,
		Payload: []byte{},
	})
	return
}

func SendChokeMessage(peer *models.Peer) (err error) {
	_, err = SendMessageWithRetry(peer, models.Message{
		Type:    models.MsgTypeChoke,
		Payload: []byte{},
	})
	return
}

func SendMessageWithRetry(peer *models.Peer, message models.Message) (int, error) {
	retries := 0
	var err error = nil

	for retries < 10 {
		_, err := peer.Conn.Write(message.ToBytes())
		if (err == nil || err == io.EOF) && retries == 0 {
			return retries, err
		}

		retries++
	}

	return retries, err
}

func ReadMessage(reader io.Reader) (*models.Message, error) {
	var lengthBytes [4]byte
	_, err := io.ReadFull(reader, lengthBytes[:])

	if err != nil {
		if err == io.EOF {
			fmt.Printf("EOF while reading message length\n")
			return nil, err
		}
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBytes[:])

	if length == 0 {
		return nil, nil
	}

	var payload []byte

	if length > 0 {
		payload = make([]byte, length)
		_, err := io.ReadFull(reader, payload)

		if err != nil {
			return nil, err
		}
	}

	return &models.Message{Type: models.MessageType(payload[0]), Payload: payload[1:]}, nil
}

func ReadRequestMessage(payload []byte) (int, int, int, error) {
	if len(payload) < 12 {
		return -1, -1, -1, errors.New("invalid payload length during request message")
	}

	pieceIndex := binary.BigEndian.Uint32(payload[0:4])
	begin := binary.BigEndian.Uint32(payload[4:8])
	length := binary.BigEndian.Uint32(payload[8:])
	return int(pieceIndex), int(begin), int(length), nil
}
