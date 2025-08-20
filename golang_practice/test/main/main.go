package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
)

func udpClient(id int, serverAddr string, wg *sync.WaitGroup, sendInterval time.Duration) {
	defer wg.Done()

	conn, err := net.Dial("udp", serverAddr)
	if err != nil {
		fmt.Printf("Client %d: lỗi kết nối: %v\n", id, err)
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(sendInterval)
	defer ticker.Stop()

	msg := []byte("ping") // message đơn giản

	for {
		select {
		case <-ticker.C:
			_, err := conn.Write(msg)
			if err != nil {
				fmt.Printf("Client %d: lỗi gửi message: %v\n", id, err)
				return
			}
			// Bạn có thể in log nếu muốn debug, nhưng với nhiều client thì nên hạn chế
			// fmt.Printf("Client %d gửi message\n", id)
		}
	}
}

func main() {
	var (
		serverAddr     string
		clientCount    int
		sendIntervalMs int
	)

	flag.StringVar(&serverAddr, "server", "127.0.0.1:9999", "Địa chỉ server UDP")
	flag.IntVar(&clientCount, "clients", 100, "Số lượng client UDP giả lập")
	flag.IntVar(&sendIntervalMs, "interval", 1000, "Khoảng cách gửi message (ms) mỗi client")
	flag.Parse()

	fmt.Printf("Bắt đầu tạo %d client gửi message đến %s mỗi %d ms\n", clientCount, serverAddr, sendIntervalMs)

	var wg sync.WaitGroup
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < clientCount; i++ {
		wg.Add(1)
		go udpClient(i, serverAddr, &wg, time.Duration(sendIntervalMs)*time.Millisecond)

		// Tạo ngắt quãng nhỏ để tránh tạo quá nhanh (tùy chọn)
		time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	}

	wg.Wait()
}
