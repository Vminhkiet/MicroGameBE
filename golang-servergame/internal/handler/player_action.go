package handler

import (
	"fmt"

	"github.com/VMinhKiet/golang-servergame/pkg/packets/action"
	"github.com/VMinhKiet/golang-servergame/pkg/player"
	"google.golang.org/protobuf/proto"
)

func HandlePlayerAction(payload []byte, addr string) {
	var msg action.PlayerAction
	if err := proto.Unmarshal(payload, &msg); err != nil {
		fmt.Println("❌ Lỗi giải mã PlayerAction:", err)
		return
	}

	p, ok := player.GetByAddr(addr)
	if !ok {
		return
	}

	p.X = msg.X
	p.Y = msg.Y
	fmt.Printf("🎯 Player %d di chuyển đến (%.2f, %.2f)\n", msg.PlayerId, msg.X, msg.Y)
}
