package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

// Backend 서버 정보
type Backend struct {
	URL          *url.URL
	Alive        bool
	mu           sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
	Connections  int64
}

func (b *Backend) SetAlive(alive bool) {
	b.mu.Lock()
	b.Alive = alive
	b.mu.Unlock()
}

func (b *Backend) IsAlive() bool {
	b.mu.RLock()
	alive := b.Alive
	b.mu.RUnlock()
	return alive
}

// LoadBalancer 구조체
type LoadBalancer struct {
	backends []*Backend
	current  uint64
}

// Round Robin 방식으로 다음 서버 선택
func (lb *LoadBalancer) NextBackendRoundRobin() *Backend {
	for i := 0; i < len(lb.backends); i++ {
		idx := atomic.AddUint64(&lb.current, 1) % uint64(len(lb.backends))
		backend := lb.backends[idx]

		if backend.IsAlive() {
			return backend
		}
	}
	return nil
}

// Least Connections 방식으로 다음 서버 선택
func (lb *LoadBalancer) NextBackendLeastConn() *Backend {
	var selected *Backend
	minConn := int64(^uint64(0) >> 1) // max int64

	for _, backend := range lb.backends {
		if !backend.IsAlive() {
			continue
		}
		conn := atomic.LoadInt64(&backend.Connections)
		if conn < minConn {
			minConn = conn
			selected = backend
		}
	}

	return selected
}

// Health Check
func (lb *LoadBalancer) HealthCheck() {
	for _, backend := range lb.backends {
		go func(b *Backend) {
			client := &http.Client{Timeout: 2 * time.Second}
			resp, err := client.Get(b.URL.String() + "/health")

			if err != nil || resp.StatusCode != http.StatusOK {
				log.Printf("❌ [Health Check] %s is DOWN", b.URL)
				b.SetAlive(false)
			} else {
				if !b.IsAlive() {
					log.Printf("✅ [Health Check] %s is UP", b.URL)
				}
				b.SetAlive(true)
			}

			if resp != nil {
				resp.Body.Close()
			}
		}(backend)
	}
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Least Connection 알고리즘 사용
	backend := lb.NextBackendLeastConn()

	if backend == nil {
		http.Error(w, "서비스를 사용할 수 없습니다", http.StatusServiceUnavailable)
		return
	}

	atomic.AddInt64(&backend.Connections, 1)
	defer atomic.AddInt64(&backend.Connections, -1)

	log.Printf("📤 요청을 %s 로 전달합니다", backend.URL)
	backend.ReverseProxy.ServeHTTP(w, r)
}

// 백엔드 서버 생성
func createBackendServer(port int) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // 처리 시간 시뮬레이션
		fmt.Fprintf(w, "응답: 서버 포트 %d가 요청을 처리했습니다!\n", port)
		log.Printf("✅ [서버:%d] 요청 처리 완료", port)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	addr := fmt.Sprintf(":=%d", port)
	log.Printf("🚀 백엔드 서버 시작: 포트 %d", port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("서버 %d 시작 실패: %v", port, err)
	}
}

func main() {
	// 백엔드 서버 포트들
	backendPorts := []int{8081, 8082, 8083}

	// 백엔드 서버들을 고루틴으로 시작
	for _, port := range backendPorts {
		go createBackendServer(port)
	}

	// 잠시 대기 (서버들이 시작될 시간)
	time.Sleep(1 * time.Second)

	// 로드 밸런서 설정
	lb := &LoadBalancer{
		backends: make([]*Backend, 0),
	}

	for _, port := range backendPorts {
		serverURL, _ := url.Parse(fmt.Sprintf("http://localhost:%d", port))
		proxy := httputil.NewSingleHostReverseProxy(serverURL)

		backend := &Backend{
			URL:          serverURL,
			Alive:        true,
			ReverseProxy: proxy,
			Connections:  0,
		}

		lb.backends = append(lb.backends, backend)
		log.Printf("✅ 백엔드 추가: %s", serverURL)
	}

	// Health Check 주기적 실행
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			lb.HealthCheck()
		}
	}()

	// 초기 Health Check
	lb.HealthCheck()
	time.Sleep(500 * time.Millisecond)

	// 로드 밸런서 시작
	lbPort := ":8080"
	log.Printf("🎯 로드 밸런서 시작: 포트 %s", lbPort)
	log.Printf("📍 접속: http://localhost%s", lbPort)

	if err := http.ListenAndServe(lbPort, lb); err != nil {
		log.Fatalf("로드 밸런서 시작 실패: %v", err)
	}
}
