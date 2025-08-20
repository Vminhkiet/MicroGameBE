package player

import "net"

type Player struct {
	ID       uint32
	Nickname string
	Team     uint32
	Addr     *net.UDPAddr
	LastSeen int64

	X  float32
	Y  float32
	HP int32
}
