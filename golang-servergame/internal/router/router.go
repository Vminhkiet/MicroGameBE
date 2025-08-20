package router

import (
	"fmt"

	"github.com/VMinhKiet/golang-servergame/pkg/packets/shared"
	"google.golang.org/protobuf/proto"
)

type HanlderFunc func([]byte, string)

var handlers = make(map[shared.PacketType]HanlderFunc)

func Register(t shared.PacketType, h HanlderFunc) {
	handlers[t] = h
}

func RoutePacket(payload []byte, addr string) {
	var env shared.Envelope

	if err := proto.Unmarshal(payload, &env); err != nil {
		fmt.Println("❌ Envelope decode failed:", err)
		return
	}

	handler, ok := handlers[env.Type]

	if !ok {
		fmt.Println("⚠️ No handler for PacketType:", env.Type)
		return
	}

	handler(env.Payload, addr)

}
