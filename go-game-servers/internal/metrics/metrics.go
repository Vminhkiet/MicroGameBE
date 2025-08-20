package metrics

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	PacketCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "udp_packets_received_total",
			Help: "Tổng số packet UDP nhận được.",
		},
	)
)

func init() {
	prometheus.MustRegister(PacketCount)
}

func StartMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	log.Println("Bắt đầu expose metrics tại :2112/metrics")
	err := http.ListenAndServe(":2112", nil)
	if err != nil {
		log.Fatalf("Không thể khởi động metrics server: %v", err)
	}
}
