package common

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"torrentClient/models"

	"github.com/IncSW/go-bencode"
)

func GetPeersList(manifest models.Manifest, peerId [20]byte, port int) (peers []models.PeerAddress, err error) {
	fmt.Printf("Getting peers list from trackers\n")
	announcer := []string{manifest.Announce}
	announcer = append(announcer, manifest.AnnounceList...)
	var trackerResponse interface{} = nil

	for _, announce := range announcer {
		announceUrl, err := getTrackerRequestUrl(manifest, announce, peerId, port)
		if err != nil {
			continue
		}
		response, err := getTrackerResponse(announceUrl)

		if err != nil {
			continue
		}
		fmt.Printf("Got peers list from tracker %v\n", announce)
		trackerResponse = response
		break
	}

	if trackerResponse == nil {
		return nil, errors.New("can't get peers from any tracker")
	}
	peers, err = getPeersFromTrackerResponse(trackerResponse)
	return
}

func getPeersFromTrackerResponse(trackerResponse interface{}) (peers []models.PeerAddress, err error) {
	receivedPeers := trackerResponse.(map[string]interface{})["peers"].([]byte)

	if len(receivedPeers)%6 != 0 {
		return nil, errors.New("invalid peers list")
	}

	for i := 0; i < len(receivedPeers); i += 6 {
		peers = append(peers, models.PeerAddress{
			IP:   receivedPeers[i : i+4],
			Port: binary.BigEndian.Uint16(receivedPeers[i+4 : i+6]),
		})
	}

	return
}

func getTrackerResponse(announceUrl string) (trackerResp interface{}, err error) {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(announceUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data := make([]byte, 2048)
	resp.Body.Read(data)

	trackerResp, err = bencode.Unmarshal(data)
	return
}

func getTrackerRequestUrl(manifest models.Manifest, announce string, peerId [20]byte, port int) (string, error) {
	baseUrl, err := url.Parse(announce)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"info_hash":  []string{string(manifest.InfoHash[:])},
		"peer_id":    []string{string(peerId[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(int(manifest.Length))},
	}

	baseUrl.RawQuery = params.Encode()
	return baseUrl.String(), nil
}

func ConnectToPeer(peerAddress models.PeerAddress, port int, timeout time.Duration) (conn net.Conn, err error) {
	return net.DialTimeout("tcp", peerAddress.IP.String()+":"+strconv.Itoa(int(peerAddress.Port)), timeout)
}

func EstablishConnection(peerAddress models.PeerAddress, manifest models.Manifest) (peer *models.Peer) {
	var conn net.Conn = nil
	var timeout = time.Duration(10 * time.Second)

	fmt.Printf("Connecting to peer %v:%v\n", peerAddress.IP, peerAddress.Port)

	for conn == nil {
		timeout *= 2
		conn, _ = ConnectToPeer(peerAddress, Port, timeout)

		if timeout > time.Duration(60*time.Second) {
			fmt.Printf("Can't connect to peer %v:%v\n", peerAddress.IP, peerAddress.Port)
			return nil
		}
	}

	fmt.Printf("Connected to peer %v:%v\n", peerAddress.IP, peerAddress.Port)

	peer = &models.Peer{
		Conn:       conn,
		Address:    peerAddress,
		Interested: false,
		IsChoked:   true,
		IsChoking:  false,
		BitField:   make([]byte, len(manifest.PieceHashes)),
	}
	return peer
}
