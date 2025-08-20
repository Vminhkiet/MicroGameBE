package client

import (
	"net"
	"time"
)

type Client struct {
	Addr *net.UDPAddr
	PlayerID uint32
	LastSeen time.Time
}