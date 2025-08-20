package gameloop

import (
	"log"
	"net"
	"time"

	"github.com/VMinhKiet/golang-servergame/pkg/packets/game"
	"github.com/VMinhKiet/golang-servergame/pkg/packets/shared"
	"github.com/VMinhKiet/golang-servergame/pkg/player"
	"google.golang.org/protobuf/proto"
)

const TICK_RATE = 60
const TICK_DURATION = time.Second / TICK_RATE

var currentTick int32 = 0

func Start(conn *net.UDPConn) {
	ticker := time.NewTicker(TICK_DURATION)

	defer ticker.Stop()

	for range ticker.C {
		currentTick++
		log.Printf("🕒 Tick #%d", currentTick)
		updateGameLogic()

		broadcastGameState(conn)
	}

}

func updateGameLogic() {

}

func broadcastGameState(conn *net.UDPConn) {
	players := player.GetAll()

	update := &game.GameStateUpdate{
		Tick: currentTick,
	}

	for _, p := range players {
		update.Players = append(update.Players, &game.PlayerState{
			PlayerId: p.ID,
			X:        p.X,
			Y:        p.Y,
			Hp:       p.HP,
			IsDead:   p.HP <= 0,
		})
	}

	payload, err := proto.Marshal(update)
	if err != nil {
		return
	}

	env := &shared.Envelope{
		Type:    shared.PacketType_GAME_STATE_UPDATE,
		Payload: payload,
	}

	packet, err := proto.Marshal(env)
	if err != nil {
		return
	}

	for _, p := range players {
		conn.WriteToUDP(packet, p.Addr)
	}

}
