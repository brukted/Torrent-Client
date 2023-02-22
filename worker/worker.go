package worker

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
	"torrentClient/common"
	"torrentClient/models"
	"torrentClient/seed"
)

func processIncomingMessages(peer *models.Peer, connReader io.Reader, progress *models.PieceJobProgress, seedRequestChannel *chan *seed.SeedRequest) (models.MessageType, error) {
	message, err := common.ReadMessage(connReader)

	if err != nil {
		fmt.Printf("Error reading message from peer %v:%v, %v\n", peer.Address.IP, peer.Address.Port, err)
		return models.MsgTypeKeepAlive, err
	}

	if message == nil {
		return models.MsgTypeKeepAlive, nil
	}

	if message.Type != models.MsgTypePiece {
		fmt.Printf("Received message from peer %v:%v, %v\n", peer.Address.IP, peer.Address.Port, message.Type.String())
	}

	switch message.Type {
	case models.MsgTypeUnChoke:
		peer.IsChoking = false
	case models.MsgTypeChoke:
		peer.IsChoking = true
	case models.MsgTypeInterested:
		peer.Interested = true
	case models.MsgTypeNotInterested:
		peer.Interested = false
	case models.MsgTypeHave:
		pieceIndex := binary.BigEndian.Uint32(message.Payload)
		peer.BitField.MarkPiece(int(pieceIndex))
	case models.MsgTypeBitField:
		peer.BitField = message.Payload
	case models.MsgTypeCancel:
		fmt.Printf("Received cancel message from peer %v:%v\n", peer.Address.IP, peer.Address.Port)
	case models.MsgTypePiece:
		index, begin, block, err := common.ReadPieceMessage(message.Payload)
		if err != nil {
			fmt.Printf("Error reading piece job result from peer %v:%v, %v\n", peer.Address.IP, peer.Address.Port, err)
			return models.MsgTypePiece, err
		}

		if progress == nil {
			fmt.Printf("Received piece job result from peer %v:%v with no job in progress\n", peer.Address.IP, peer.Address.Port)
			return models.MsgTypePiece, err
		}

		if index != progress.PieceIndex {
			fmt.Printf("Received piece job result from peer %v:%v with wrong piece index %v\n", peer.Address.IP, peer.Address.Port, index)
			return models.MsgTypePiece, err
		}

		if begin+len(block) > progress.PieceLength {
			fmt.Printf("Received piece job result from peer %v:%v with wrong begin %v\n", peer.Address.IP, peer.Address.Port, begin)
			return models.MsgTypePiece, err
		}

		progress.TotalDownloaded += len(block)
		copy(progress.Buffer[begin:], block)
	case models.MsgTypeRequest:
		*seedRequestChannel <- &seed.SeedRequest{
			Peer:    peer,
			Message: message,
		}
		return models.MsgTypeRequest, nil
	}

	return message.Type, nil
}

func readHandShake(connReader io.Reader, peer *models.Peer, manifest models.Manifest) bool {
	handshake, err := common.ReadHandShake(connReader)
	if err != nil {
		fmt.Printf("Error reading handshake from peer %v:%v, %v\n", peer.Address.IP, peer.Address.Port, err)
		return true
	}

	if !bytes.Equal(handshake.InfoHash[:], manifest.InfoHash[:]) {
		fmt.Printf("Handshake info hash doesn't match with manifest info hash from peer %v:%v\n", peer.Address.IP, peer.Address.Port)
		return true
	}
	fmt.Printf("Handshake established with peer %v:%v\n", peer.Address.IP, peer.Address.Port)
	return false
}

func sendChoke(peer *models.Peer) bool {
	_, err := common.SendMessageWithRetry(peer, models.Message{
		Type: models.MsgTypeChoke,
	})
	if err != nil {
		fmt.Printf("Error sending choke to peer %v:%v\n", peer.Address.IP, peer.Address.Port)
		return true
	}
	fmt.Printf("Choke sent to peer %v:%v\n", peer.Address.IP, peer.Address.Port)
	return false
}

func StartPeerWorker(peers []*models.Peer, i int, peerAddress models.PeerAddress, peerId [20]byte, manifest models.Manifest, port int, workChannel *chan models.PieceJob, currentBitField *models.Bitfield, pieceJobResultChannel *chan *models.PieceJobResult, seedRequestChannel *chan *seed.SeedRequest, conn *net.Conn) {
	// Establish connection
	var peer *models.Peer = nil

	if conn != nil {
		peer = &models.Peer{
			Address:    peerAddress,
			Conn:       *conn,
			Interested: false,
			IsChoking:  true,
			IsChoked:   false,
			BitField:   make(models.Bitfield, len(*currentBitField)),
		}
	} else {
		peer = common.EstablishConnection(peerAddress, manifest)
	}

	if peer == nil {
		return
	}
	peers[i] = peer
	defer peer.Conn.Close()

	// Establish handshake
	_, err := common.EstablishHandShake(peerId, *peer, manifest)
	if err != nil {
		fmt.Printf("Error establishing handshake with peer %v:%v\n", peer.Address.IP, peer.Address.Port)
		return
	}

	connReader := io.Reader(peer.Conn)

	// Read handshake
	shouldReturn := readHandShake(connReader, peer, manifest)
	if shouldReturn {
		return
	}

	// Send Interested
	_, err = common.SendMessageWithRetry(peer, models.Message{
		Type: models.MsgTypeInterested,
	})
	if err != nil {
		fmt.Printf("Error sending interested to peer %v:%v\n", peer.Address.IP, peer.Address.Port)
		return
	}
	fmt.Printf("Interested sent to peer %v:%v\n", peer.Address.IP, peer.Address.Port)

	err = common.SendUnchokeMessage(peer)
	if err != nil {
		fmt.Printf("Error sending unchoke to peer %v:%v\n", peer.Address.IP, peer.Address.Port)
		return
	}

	// Receive bitfield
	_, err = processIncomingMessages(peer, connReader, nil, seedRequestChannel)
	if err != nil {
		fmt.Printf("Error processing incoming messages from peer %v:%v, %v\n", peer.Address.IP, peer.Address.Port, err)
		return
	}

	for {
		// Check if peer is choking us and wait for unchoke
		if peer.IsChoking {
			time.Sleep(1 * time.Second)
			_, err = processIncomingMessages(peer, connReader, nil, seedRequestChannel)
			if err != nil {
				fmt.Printf("Error processing incoming messages from peer %v:%v, %v\n", peer.Address.IP, peer.Address.Port, err)
				return
			}
			continue
		}

		for pieceJob := range *workChannel {
			// Check if peer has piece
			if !peer.BitField.HasPiece(pieceJob.PieceIndex) {
				*workChannel <- pieceJob
				continue
			}

			if peer.IsChoking {
				*workChannel <- pieceJob
				break
			}

			fmt.Printf("Sending piece job to peer %v:%v, piece index %v\n", peer.Address.IP, peer.Address.Port, pieceJob.PieceIndex)

			pieceJobProgress := models.PieceJobProgress{
				PieceIndex:      pieceJob.PieceIndex,
				Buffer:          make([]byte, pieceJob.PieceLength),
				TotalDownloaded: 0,
				PieceLength:     pieceJob.PieceLength,
			}

			peer.Conn.SetDeadline(time.Now().Add(30 * time.Second))

			for pieceJobProgress.TotalDownloaded < pieceJob.PieceLength {
				if peer.IsChoking {
					*workChannel <- pieceJob
					break
				}

				currentBlockSize := common.BlockSize
				if pieceJob.PieceLength-pieceJobProgress.TotalDownloaded < common.BlockSize {
					currentBlockSize = pieceJob.PieceLength - pieceJobProgress.TotalDownloaded
				}

				// Send request
				_, err = common.SendMessageWithRetry(peer, models.Message{
					Type: models.MsgTypeRequest,
					Payload: models.RequestMessage{
						PieceIndex: pieceJob.PieceIndex,
						Begin:      pieceJobProgress.TotalDownloaded,
						Length:     currentBlockSize,
					}.ToBytes(),
				})

				if err != nil {
					fmt.Printf("Error sending request to peer %v:%v\n", peer.Address.IP, peer.Address.Port)
					*workChannel <- pieceJob
					return
				}

				var msgType models.MessageType = models.MsgTypeKeepAlive

				for msgType != models.MsgTypePiece {
					msgType, err = processIncomingMessages(peer, connReader, &pieceJobProgress, seedRequestChannel)
					if peer.IsChoking {
						break
					}

					if err != nil {
						fmt.Printf("Error processing incoming messages from peer %v:%v, %v\n", peer.Address.IP, peer.Address.Port, err)
						*workChannel <- pieceJob
						return
					}
				}
			}

			// Done with piece job
			// Reset deadline
			peer.Conn.SetDeadline(time.Time{})

			if pieceJobProgress.TotalDownloaded == pieceJob.PieceLength {
				// check if piece is valid
				if !common.CheckPieceHash(pieceJobProgress.Buffer, pieceJob.PieceHash) {
					fmt.Printf("Piece hash doesn't match for piece %v\n", pieceJob.PieceIndex)
					*workChannel <- pieceJob
					continue
				}

				*pieceJobResultChannel <- &models.PieceJobResult{
					PieceIndex: pieceJob.PieceIndex,
					PieceData:  pieceJobProgress.Buffer,
				}
			}
		}
	}
}
