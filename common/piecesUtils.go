package common

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"os"
	"torrentClient/models"
)

func GetPieceLength(index int, pieceLength int, totalLength int) int {
	if pieceLength < (totalLength - index*pieceLength) {
		return pieceLength
	}
	return totalLength - index*pieceLength
}

func ReadPieceMessage(payload []byte) (int, int, []byte, error) {
	if len(payload) < 8 {
		return 0, 0, nil, errors.New("invalid payload length during piece message")
	}
	pieceIndex := binary.BigEndian.Uint32(payload[0:4])
	begin := binary.BigEndian.Uint32(payload[4:8])
	block := payload[8:]
	return int(pieceIndex), int(begin), block, nil
}

func CheckPieceHash(piece []byte, hash [20]byte) bool {
	sha1Hash := sha1.Sum(piece)
	return bytes.Equal(sha1Hash[:], hash[:])
}

func WritePieceToFile(manifest *models.Manifest, pieceJobResult *models.PieceJobResult, blobFile *os.File) {
	pieceOffset := int64(pieceJobResult.PieceIndex) * manifest.PieceLength
	blobFile.WriteAt(pieceJobResult.PieceData, pieceOffset)
}

func WritePieceMessage(index int, begin int, block []byte) *models.Message {
	payload := make([]byte, 8)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	payload = append(payload, block...)
	return &models.Message{
		Type:    models.MsgTypePiece,
		Payload: payload,
	}
}
