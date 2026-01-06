package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

func sendRequest(client *http.Client, url string, id int, wg *sync.WaitGroup) {
	defer wg.Done()

	resp, err := client.Get(url)
	if err != nil {
		log.Printf("❌ [요청 %d] 실패: %v", id, err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("✅ [요청 %d] %s", id, string(body))
}

func main() {
	loadBalancerURL := "http://localhost:8080"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	log.Println("🔗 로드 밸런서에 연결 중...")
	time.Sleep(1 * time.Second)

	// 1. 순차 요청 테스트
	log.Println("\n=== 순차 요청 테스트 (5개) ===")
	for i := 1; i <= 5; i++ {
		resp, err := client.Get(loadBalancerURL)
		if err != nil {
			log.Printf("❌ 요청 %d 실패: %v", i, err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		log.Printf("✅ [요청 %d] %s", i, string(body))
		resp.Body.Close()
		time.Sleep(200 * time.Millisecond)
	}

	// 2. 동시 요청 테스트
	log.Println("\n=== 동시 요청 테스트 (20개) ===")
	var wg sync.WaitGroup
	start := time.Now()

	for i := 1; i <= 20; i++ {
		wg.Add(1)
		go sendRequest(client, loadBalancerURL, i, &wg)
		time.Sleep(50 * time.Millisecond) // 약간의 지연
	}

	wg.Wait()
	elapsed := time.Since(start)

	log.Printf("\n⏱️  20개 동시 요청 완료 시간: %s", elapsed)
	log.Println("✨ 모든 테스트 완료!")
	log.Println("💡 각 요청이 어느 서버에서 처리되었는지 확인해보세요!")
}
