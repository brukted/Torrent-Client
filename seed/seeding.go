package seed

import (
	"fmt"
	"torrentClient/common"
	"torrentClient/models"
)

func HandleSeedingRequest(req *SeedRequest, blobFile []byte, currentBitField *models.Bitfield, manifest *models.Manifest) {
	if req.Peer.IsChoked {
		common.SendMessageWithRetry(req.Peer, models.Message{Type: models.MsgTypeChoke})
		return
	}

	index, begin, length, err := common.ReadRequestMessage(req.Message.Payload)

	if err != nil {
		fmt.Printf("Error reading request message from peer %v:%v, %v\n", req.Peer.Address.IP, req.Peer.Address.Port, err)
		return
	}

	if index >= len(*currentBitField) {
		fmt.Printf("Received request message from peer %v:%v with invalid index %v\n", req.Peer.Address.IP, req.Peer.Address.Port, index)
		return
	}

	if !(*currentBitField).HasPiece(index) {
		fmt.Printf("Received request message from peer %v:%v with invalid index %v\n", req.Peer.Address.IP, req.Peer.Address.Port, index)
		return
	}

	if begin+length > int(manifest.PieceLength) {
		fmt.Printf("Received request message from peer %v:%v with invalid begin %v\n", req.Peer.Address.IP, req.Peer.Address.Port, begin)
		return
	}

	pieceOffset := int64(index) * int64(manifest.PieceLength)
	blockOffset := int64(begin)
	block := blobFile[pieceOffset+blockOffset : pieceOffset+blockOffset+int64(length)]

	if err != nil {
		fmt.Printf("Error reading block from blob file %v\n", err)
		return
	}

	common.SendMessageWithRetry(req.Peer, *common.WritePieceMessage(index, begin, block))
}
