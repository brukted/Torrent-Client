package models

import (
	"net"
)

type Peer struct {
	Conn       net.Conn
	Address    PeerAddress
	Interested bool
	// IsChoked is true if the peer is not allowed to send us data
	IsChoked bool
	// IsChoking is true if we are not allowed to send data to the peer
	IsChoking bool
	// BitField is a list of booleans that indicate whether the peer has the corresponding piece
	BitField Bitfield
}
