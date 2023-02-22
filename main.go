package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"torrentClient/common"
	"torrentClient/models"
	"torrentClient/seed"
	"torrentClient/worker"
)

func main() {
	manifest := common.ReadManifestFromFile("debian-11.6.0-amd64-netinst.iso.torrent")

	// Create files
	blobFile := common.LoadOrCreateDownloadBlob(&manifest)
	memCopy := make([]byte, manifest.Length)
	blobFile.ReadAt(memCopy, 0)

	// Load progress from persistent storage
	currentBitField, bitfieldFile := models.LoadOrCreateBitFieldFromFile(&manifest)
	totalDownloaded := 0

	// count already downloaded pieces
	for _, piece := range *currentBitField {
		for i := 0; i < 8; i++ {
			if piece&(1<<uint(i)) != 0 {
				totalDownloaded++
			}
		}
	}

	fmt.Println("Total downloaded", totalDownloaded)

	// Get peers list
	id := [20]byte{}
	rand.Read(id[:])

	peerAddresses, err := common.GetPeersList(manifest, id, common.Port)
	if err != nil {
		fmt.Println("Can't get peers", err)
		panic(err)
	}
	fmt.Println(peerAddresses)

	// channels
	workChannel := make(chan models.PieceJob, len(manifest.PieceHashes))
	pieceJobResultChannel := make(chan *models.PieceJobResult)
	seedRequestChannel := make(chan *seed.SeedRequest)

	// create work for each piece
	for index, hash := range manifest.PieceHashes {
		// ignore already downloaded pieces
		if !currentBitField.HasPiece(index) {
			workChannel <- models.PieceJob{
				PieceIndex:  index,
				PieceHash:   hash,
				PieceLength: common.GetPieceLength(index, int(manifest.PieceLength), int(manifest.Length)),
			}
		}
	}

	// create common structure for leecher and seeder
	peers := make([]*models.Peer, len(peerAddresses))

	for i, peerAddress := range peerAddresses {
		go worker.StartPeerWorker(peers, i, peerAddress, id, manifest, common.Port, &workChannel, currentBitField, &pieceJobResultChannel, &seedRequestChannel, nil)
	}

	// Listen for seeding requests
	go func() {
		for {
			seedRequest := <-seedRequestChannel
			// handle seeding request
			go seed.HandleSeedingRequest(seedRequest, memCopy, currentBitField, &manifest)
		}
	}()

	// Start seeding server
	go func() {
		ListenAddr := ":" + fmt.Sprint(common.Port)
		listener, err := net.Listen("tcp", ListenAddr)
		if err != nil {
			log.Fatal(err)
		}

		defer listener.Close()

		log.Printf("Listening on %s...\n", ListenAddr)

		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println(err)
				continue
			}
			peers := append(peers, nil)
			addr := models.PeerAddress{
				IP:   conn.RemoteAddr().(*net.TCPAddr).IP,
				Port: uint16(common.Port),
			}

			go worker.StartPeerWorker(peers, len(peers)-1, addr, id, manifest, common.Port, &workChannel, currentBitField, &pieceJobResultChannel, &seedRequestChannel, &conn)
		}
	}()

	// Optimistic Unchoking
	go func() {
		for {
			if len(peers) != 0 {
				// unchoke random peer
				peerIndex := rand.Intn(len(peers))
				if peers[peerIndex] != nil {
					if peers[peerIndex].IsChoked {
						peers[peerIndex].IsChoked = false
						go common.SendUnchokeMessage(peers[peerIndex])
					}
				}
			}
			time.Sleep(31 * time.Second)
		}
	}()

	// process results
	for {
		pieceJobResult := <-pieceJobResultChannel
		if pieceJobResult == nil {
			continue
		}

		copy(memCopy[pieceJobResult.PieceIndex*int(manifest.PieceLength):], pieceJobResult.PieceData)
		// write piece to file
		common.WritePieceToFile(&manifest, pieceJobResult, blobFile)

		// update bitfield
		currentBitField.MarkPiece(pieceJobResult.PieceIndex)
		currentBitField.WriteToFile(&manifest, bitfieldFile)

		// update progress
		totalDownloaded++
		fmt.Printf("Downloaded %v/%v pieces\n", totalDownloaded, len(manifest.PieceHashes))

		// send have message to all peers
		for _, peer := range peers {
			if peer != nil {
				common.SendHaveMessage(peer, pieceJobResult.PieceIndex)
			}
		}

		// check if download is finished
		if totalDownloaded == len(manifest.PieceHashes) {
			fmt.Println("Download finished")
			common.WriteBlobToFiles(&manifest)
		}
	}
}
