package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// TokenBucket Rate Limiter
type TokenBucket struct {
	capacity      int           // 버킷 용량
	tokens        int           // 현재 토큰 수
	refillRate    int           // 초당 리필되는 토큰 수
	lastRefillTime time.Time
	mu            sync.Mutex
}

func NewTokenBucket(capacity int, refillRate int) *TokenBucket {
	return &TokenBucket{
		capacity:       capacity,
		tokens:         capacity,
		refillRate:     refillRate,
		lastRefillTime: time.Now(),
	}
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime)

	// 경과 시간에 따라 토큰 추가
	tokensToAdd := int(elapsed.Seconds() * float64(tb.refillRate))
	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefillTime = now
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

func (tb *TokenBucket) GetTokens() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.refill()
	return tb.tokens
}

// Fixed Window Rate Limiter
type FixedWindow struct {
	limit      int
	window     time.Duration
	requests   int
	windowStart time.Time
	mu         sync.Mutex
}

func NewFixedWindow(limit int, window time.Duration) *FixedWindow {
	return &FixedWindow{
		limit:       limit,
		window:      window,
		requests:    0,
		windowStart: time.Now(),
	}
}

func (fw *FixedWindow) Allow() bool {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	now := time.Now()

	// 새로운 윈도우 시작
	if now.Sub(fw.windowStart) >= fw.window {
		fw.requests = 0
		fw.windowStart = now
	}

	// 요청 수 체크
	if fw.requests < fw.limit {
		fw.requests++
		return true
	}

	return false
}

func (fw *FixedWindow) GetRemaining() int {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	now := time.Now()
	if now.Sub(fw.windowStart) >= fw.window {
		return fw.limit
	}

	return fw.limit - fw.requests
}

// 사용자별 Rate Limiter
type UserRateLimiter struct {
	limiters map[string]*TokenBucket
	mu       sync.Mutex
	capacity int
	refillRate int
}

func NewUserRateLimiter(capacity int, refillRate int) *UserRateLimiter {
	return &UserRateLimiter{
		limiters:   make(map[string]*TokenBucket),
		capacity:   capacity,
		refillRate: refillRate,
	}
}

func (url *UserRateLimiter) Allow(userID string) bool {
	url.mu.Lock()
	limiter, exists := url.limiters[userID]
	if !exists {
		limiter = NewTokenBucket(url.capacity, url.refillRate)
		url.limiters[userID] = limiter
	}
	url.mu.Unlock()

	return limiter.Allow()
}

// API 요청 시뮬레이션
func makeAPIRequest(limiter *TokenBucket, requestID int) {
	if limiter.Allow() {
		log.Printf("✅ [요청 %d] 허용됨 (남은 토큰: %d)", requestID, limiter.GetTokens())
	} else {
		log.Printf("🚫 [요청 %d] 거부됨 (Rate Limit 초과)", requestID)
	}
}

func main() {
	log.Println("🚀 Rate Limiting 시스템 시작\n")

	// 시나리오 1: Token Bucket - 정상 사용
	log.Println("=== 시나리오 1: Token Bucket (용량: 5, 리필: 2/초) ===")
	log.Println("💡 초당 2개씩 토큰이 리필됩니다\n")

	tb := NewTokenBucket(5, 2)

	log.Println("초기 상태:")
	for i := 1; i <= 5; i++ {
		makeAPIRequest(tb, i)
		time.Sleep(100 * time.Millisecond)
	}

	log.Println("\n토큰 소진 후 추가 요청:")
	for i := 6; i <= 8; i++ {
		makeAPIRequest(tb, i)
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("\n⏳ 1초 대기 (토큰 리필)...\n")
	time.Sleep(1 * time.Second)

	log.Println("리필 후 요청:")
	for i := 9; i <= 11; i++ {
		makeAPIRequest(tb, i)
		time.Sleep(100 * time.Millisecond)
	}

	// 시나리오 2: Fixed Window
	log.Println("\n=== 시나리오 2: Fixed Window (5 req/2초) ===")
	log.Println("💡 2초 윈도우 내에서 최대 5개 요청만 허용\n")

	fw := NewFixedWindow(5, 2*time.Second)

	for i := 1; i <= 8; i++ {
		if fw.Allow() {
			log.Printf("✅ [요청 %d] 허용됨 (남은: %d)", i, fw.GetRemaining())
		} else {
			log.Printf("🚫 [요청 %d] 거부됨 (윈도우 리셋까지 대기 필요)", i)
		}
		time.Sleep(200 * time.Millisecond)
	}

	log.Printf("\n⏳ 2초 대기 (윈도우 리셋)...\n")
	time.Sleep(2 * time.Second)

	log.Println("새 윈도우에서 요청:")
	for i := 9; i <= 11; i++ {
		if fw.Allow() {
			log.Printf("✅ [요청 %d] 허용됨 (남은: %d)", i, fw.GetRemaining())
		}
		time.Sleep(200 * time.Millisecond)
	}

	// 시나리오 3: 사용자별 Rate Limiting
	log.Println("\n=== 시나리오 3: 사용자별 독립적인 Rate Limiting ===")
	log.Println("💡 각 사용자는 독립적인 토큰 버킷을 가집니다 (용량: 3, 리필: 1/초)\n")

	userLimiter := NewUserRateLimiter(3, 1)

	users := []string{"user1", "user2", "user3"}

	for i := 1; i <= 5; i++ {
		for _, user := range users {
			if userLimiter.Allow(user) {
				log.Printf("✅ [%s] 요청 %d 허용됨", user, i)
			} else {
				log.Printf("🚫 [%s] 요청 %d 거부됨", user, i)
			}
		}
		time.Sleep(200 * time.Millisecond)
	}

	// 시나리오 4: 급증 트래픽 (Burst Traffic)
	log.Println("\n=== 시나리오 4: 급증 트래픽 처리 ===")
	log.Println("💡 짧은 시간에 많은 요청이 들어오는 상황\n")

	burstLimiter := NewTokenBucket(10, 5) // 용량 10, 초당 5개 리필

	log.Println("🔥 100개 요청 급증:")
	allowed := 0
	denied := 0

	start := time.Now()
	for i := 1; i <= 100; i++ {
		if burstLimiter.Allow() {
			allowed++
		} else {
			denied++
		}
		time.Sleep(10 * time.Millisecond)
	}
	elapsed := time.Since(start)

	log.Printf("\n📊 결과:")
	log.Printf("   ✅ 허용: %d개", allowed)
	log.Printf("   🚫 거부: %d개", denied)
	log.Printf("   ⏱️  처리 시간: %s", elapsed)
	log.Printf("   📈 처리율: %.1f req/s", float64(allowed)/elapsed.Seconds())

	// 시나리오 5: 동시 요청
	log.Println("\n=== 시나리오 5: 동시 요청 (Thread-Safe 테스트) ===")
	log.Println("💡 여러 goroutine이 동시에 요청을 보냅니다\n")

	concurrentLimiter := NewTokenBucket(20, 10)
	var wg sync.WaitGroup
	allowedCount := 0
	deniedCount := 0
	var countMu sync.Mutex

	for i := 1; i <= 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if concurrentLimiter.Allow() {
				countMu.Lock()
				allowedCount++
				countMu.Unlock()
			} else {
				countMu.Lock()
				deniedCount++
				countMu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	log.Printf("📊 동시 요청 결과:")
	log.Printf("   ✅ 허용: %d개", allowedCount)
	log.Printf("   🚫 거부: %d개", deniedCount)

	log.Println("\n✨ Rate Limiting 시뮬레이션 완료!")
	log.Println("💡 학습 포인트:")
	log.Println("   1. Token Bucket: 버스트 트래픽 허용, 평균 제한")
	log.Println("   2. Fixed Window: 구현 간단, 경계 문제 있음")
	log.Println("   3. 사용자별 독립적인 제한 가능")
	log.Println("   4. Thread-safe 구현 필요")
	log.Println("   5. 실제 환경: Redis, Nginx Rate Limiting 사용")
	log.Println("\n💡 비교:")
	log.Println("   - Token Bucket: 버스트 허용, 유연함")
	log.Println("   - Fixed Window: 구현 간단, 윈도우 경계 문제")
	log.Println("   - Sliding Window: 더 정확, 계산 복잡")
	log.Println("   - Leaky Bucket: 일정한 속도, 버스트 불가")
}
