package handler

import (
	"fmt"
	"log"

	"github.com/VMinhKiet/golang-servergame/client"
	"github.com/VMinhKiet/golang-servergame/pkg/packets/match"
	"github.com/VMinhKiet/golang-servergame/pkg/player"
	"google.golang.org/protobuf/proto"
)

func HandleJoinMatch(payload []byte, addr string) {
	var req match.JoinMatchRequest

	if err := proto.Unmarshal(payload, &req); err != nil {
		fmt.Println("❌ JoinMatchRequest lỗi:", err)
		return
	}

	c, ok := client.Get(addr)

	if !ok {
		log.Println("⚠️ Không tìm thấy client cho địa chỉ:", addr)
		return
	}

	player.UpdateorCreate(c.Addr, req.PlayerId, req.Nickname)

	fmt.Printf("✅ Player %s (ID: %d) đã tham gia từ %s\n", req.Nickname, req.PlayerId, addr)
}
