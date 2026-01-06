package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

var requestCount int64

type Request struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

type Response struct {
	Status    string    `json:"status"`
	Data      string    `json:"data"`
	Timestamp time.Time `json:"timestamp"`
	RequestID int64     `json:"request_id"`
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// 요청 카운트 증가
	reqID := atomic.AddInt64(&requestCount, 1)

	log.Printf("[요청 #%d] %s %s from %s", reqID, r.Method, r.URL.Path, r.RemoteAddr)

	if r.Method != http.MethodPost {
		http.Error(w, "POST 메서드만 지원합니다", http.StatusMethodNotAllowed)
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "잘못된 요청 형식입니다", http.StatusBadRequest)
		return
	}

	// 요청 처리 시뮬레이션 (약간의 지연)
	time.Sleep(100 * time.Millisecond)

	resp := Response{
		Status:    "success",
		Data:      fmt.Sprintf("안녕하세요 %s님! 메시지 '%s'를 받았습니다.", req.Name, req.Message),
		Timestamp: time.Now(),
		RequestID: reqID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	log.Printf("[응답 #%d] 처리 완료", reqID)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "healthy",
		"uptime":        time.Now().Format(time.RFC3339),
		"total_requests": atomic.LoadInt64(&requestCount),
	})
}

func main() {
	http.HandleFunc("/api/request", handleRequest)
	http.HandleFunc("/health", handleHealth)

	port := ":8080"
	log.Printf("🚀 서버가 포트 %s 에서 시작되었습니다...", port)
	log.Printf("📍 엔드포인트: http://localhost%s/api/request", port)
	log.Printf("💚 Health Check: http://localhost%s/health", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("서버 시작 실패: %v", err)
	}
}
