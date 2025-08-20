package handler

import (
	"fmt"
	"golang_practice/pkg/packets/ping"
	"golang_practice/pkg/packets/shared"
	"net"
	"time"

	"google.golang.org/protobuf/proto"
)

func HandlePingRequest(payload []byte, addr *net.UDPAddr, conn *net.UDPConn) {
	var req ping.PingRequest
	if err := proto.Unmarshal(payload, &req); err != nil {
		fmt.Printf("Lỗi giải mã PingRequest: %v\n", err)
		return
	}

	fmt.Printf("Nhận PingRequest từ %v - ClientTime: %v\n", addr, req.ClientTime)

	resp := &ping.PingResponse{
		ClientTime: req.ClientTime,
		ServerTime: time.Now().UnixNano(),
	}

	respData, err := proto.Marshal(resp)
	if err != nil {
		fmt.Printf("Lỗi mã hóa PingResponse: %v\n", err)
		return
	}

	respEnv := &shared.Envelope{
		Type:    shared.PacketType_PING_RESPONSE,
		Payload: respData,
	}

	finalData, err := proto.Marshal(respEnv)
	if err != nil {
		fmt.Printf("Lỗi mã hóa Envelope phản hồi: %v\n", err)
		return
	}

	if _, err := conn.WriteToUDP(finalData, addr); err != nil {
		fmt.Printf("Lỗi gửi phản hồi: %v\n", err)
	} else {
		fmt.Println("Đã gửi PingResponse.")
	}
}
