package router

import (
	"fmt"
	"golang_practice/pkg/packets/shared"
	"net"

	"google.golang.org/protobuf/proto"
)

type HanlderFunc func([]byte, *net.UDPAddr, *net.UDPConn)

var handlers = make(map[shared.PacketType]HanlderFunc)

func Register(t shared.PacketType, h HanlderFunc) {
	handlers[t] = h
}

func RoutePacket(conn *net.UDPConn, payload []byte, addr *net.UDPAddr) {
	var env shared.Envelope

	err := proto.Unmarshal(payload, &env)

	if err != nil {
		fmt.Printf("Loi decode %v\n", err)
		return
	}

	handler, ok := handlers[env.Type]

	if !ok {
		fmt.Println("⚠️ No handler for PacketType:", env.Type)
		return
	}

	handler(env.Payload, addr, conn)
}
