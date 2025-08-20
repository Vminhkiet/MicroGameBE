package handler

import (
	"fmt"

	"github.com/VMinhKiet/golang-servergame/pkg/packets/ping"
	"google.golang.org/protobuf/proto"
)

func HandlePing(payload []byte, addr string) {
	var req ping.PingRequest

	if err := proto.Unmarshal(payload, &req); err != nil {
		fmt.Println("❌ PingRequest lỗi:", err)
		return
	}

	fmt.Printf("📶 Ping từ %s: clientTime=%d\n", addr, req.ClientTime)
}
