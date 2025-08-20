package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Vminhkiet/BattleGround-backend/go-game-server/internal/metrics"
	"github.com/Vminhkiet/BattleGround-backend/go-game-server/internal/server" // Đảm bảo đường dẫn import đúng
)

func main() {
	go metrics.StartMetricsServer()
	serverPort := 8080 // Cổng mà server UDP sẽ lắng nghe

	srv, err := server.NewUDPServer(serverPort)
	if err != nil {
		log.Fatalf("Không thể tạo máy chủ UDP: %v", err)
	}

	err = srv.Start()
	if err != nil {
		log.Fatalf("Không thể khởi động máy chủ UDP: %v", err)
	}

	// Chờ tín hiệu dừng từ hệ điều hành (Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Dừng server một cách duyên dáng
	srv.Stop()
	log.Println("Máy chủ đã tắt.")
}
