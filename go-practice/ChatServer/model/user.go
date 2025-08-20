package model

import (
	"ChatServer/enum"
	"net"
)

type Client struct {
	Conn     net.Conn
	Username string
	Role     enum.Role
	Muted    bool
	Kicked   bool
}
