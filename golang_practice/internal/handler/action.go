package handler

import (
	"fmt"
	"golang_practice/internal/gameloop"
	"golang_practice/pkg/packets/action"
	"net"

	"google.golang.org/protobuf/proto"
)

func HandlePlayerAction(payload []byte, addr *net.UDPAddr, conn *net.UDPConn) {
	var env action.Player_Action

	err := proto.Unmarshal(payload, &env)

	if err != nil {
		fmt.Printf("Loi giai ma payload cua playerAction %v\n", err)
		return
	}

	gameloop.ApplyPlayerAction(env.PlayerId, &env, addr)
	fmt.Printf("Nhan duoc goi tin %v\n", env.PlayerId)
}
