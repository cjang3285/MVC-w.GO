package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

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

func sendRequest(client *http.Client, serverURL string, name string, message string) (*Response, error) {
	req := Request{
		Name:    name,
		Message: message,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("JSON 마샬링 실패: %w", err)
	}

	resp, err := client.Post(serverURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("요청 전송 실패: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 실패: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("서버 에러: %s - %s", resp.Status, string(body))
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("JSON 언마샬링 실패: %w", err)
	}

	return &response, nil
}

func main() {
	serverURL := "http://localhost:8080/api/request"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	log.Println("🔗 서버에 연결 중...")

	// 1. 단일 요청 테스트
	log.Println("\n=== 단일 요청 테스트 ===")
	resp, err := sendRequest(client, serverURL, "홍길동", "안녕하세요!")
	if err != nil {
		log.Fatalf("요청 실패: %v", err)
	}
	log.Printf("✅ 응답 받음 (요청 ID: %d): %s", resp.RequestID, resp.Data)
	log.Printf("   타임스탬프: %s", resp.Timestamp.Format(time.RFC3339))

	// 2. 동시 요청 테스트
	log.Println("\n=== 동시 요청 테스트 (10개) ===")
	var wg sync.WaitGroup
	start := time.Now()

	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			resp, err := sendRequest(
				client,
				serverURL,
				fmt.Sprintf("사용자%d", id),
				fmt.Sprintf("메시지 번호 %d", id),
			)

			if err != nil {
				log.Printf("❌ [클라이언트 %d] 요청 실패: %v", id, err)
				return
			}

			log.Printf("✅ [클라이언트 %d] 요청 ID %d - 응답: %s",
				id, resp.RequestID, resp.Data)
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)
	log.Printf("\n⏱️  10개 동시 요청 완료 시간: %s", elapsed)

	// 3. Health Check
	log.Println("\n=== Health Check ===")
	healthResp, err := client.Get("http://localhost:8080/health")
	if err != nil {
		log.Printf("❌ Health check 실패: %v", err)
	} else {
		defer healthResp.Body.Close()
		var health map[string]interface{}
		json.NewDecoder(healthResp.Body).Decode(&health)
		log.Printf("💚 서버 상태: %v", health)
	}

	log.Println("\n✨ 모든 테스트 완료!")
}
