package models

import "net"

type PeerAddress struct {
	IP   net.IP
	Port uint16
}
