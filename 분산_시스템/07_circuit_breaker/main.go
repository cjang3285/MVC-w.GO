package main

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

var (
	ErrCircuitOpen = errors.New("circuit breaker is OPEN")
)

type CircuitBreaker struct {
	name            string
	state           State
	failureCount    int
	successCount    int
	requestCount    int
	failureThreshold int           // 실패 임계값
	successThreshold int           // 복구를 위한 성공 임계값
	timeout         time.Duration // Open 상태 유지 시간
	lastFailTime    time.Time
	mu              sync.Mutex
}

func NewCircuitBreaker(name string, failureThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:             name,
		state:            StateClosed,
		failureThreshold: failureThreshold,
		successThreshold: 2, // Half-Open에서 2번 성공하면 Closed로 전환
		timeout:          timeout,
	}
}

// Call: 함수 실행 (Circuit Breaker 적용)
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()

	// Open 상태이고 타임아웃이 지났으면 Half-Open으로 전환
	if cb.state == StateOpen {
		if time.Since(cb.lastFailTime) > cb.timeout {
			cb.state = StateHalfOpen
			cb.failureCount = 0
			cb.successCount = 0
			log.Printf("🟡 [%s] 상태 전환: OPEN → HALF_OPEN (복구 시도)", cb.name)
		} else {
			cb.mu.Unlock()
			log.Printf("🔴 [%s] 요청 차단 (Circuit OPEN)", cb.name)
			return ErrCircuitOpen
		}
	}

	cb.requestCount++
	cb.mu.Unlock()

	// 함수 실행
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
		return err
	}

	cb.onSuccess()
	return nil
}

func (cb *CircuitBreaker) onSuccess() {
	cb.successCount++

	switch cb.state {
	case StateClosed:
		cb.failureCount = 0
	case StateHalfOpen:
		// Half-Open에서 연속 성공하면 Closed로 전환
		if cb.successCount >= cb.successThreshold {
			cb.state = StateClosed
			cb.failureCount = 0
			cb.successCount = 0
			log.Printf("🟢 [%s] 상태 전환: HALF_OPEN → CLOSED (복구 완료!)", cb.name)
		}
	}
}

func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailTime = time.Now()

	switch cb.state {
	case StateClosed:
		// Closed에서 실패가 임계값을 넘으면 Open으로 전환
		if cb.failureCount >= cb.failureThreshold {
			cb.state = StateOpen
			log.Printf("🔴 [%s] 상태 전환: CLOSED → OPEN (실패 %d회)", cb.name, cb.failureCount)
		} else {
			log.Printf("⚠️  [%s] 실패 카운트: %d/%d", cb.name, cb.failureCount, cb.failureThreshold)
		}
	case StateHalfOpen:
		// Half-Open에서 실패하면 다시 Open으로 전환
		cb.state = StateOpen
		cb.successCount = 0
		log.Printf("🔴 [%s] 상태 전환: HALF_OPEN → OPEN (복구 실패)", cb.name)
	}
}

func (cb *CircuitBreaker) GetState() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

func (cb *CircuitBreaker) GetStats() (state State, failures int, successes int, requests int) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state, cb.failureCount, cb.successCount, cb.requestCount
}

// 불안정한 서비스 시뮬레이션
type UnstableService struct {
	isHealthy bool
	mu        sync.Mutex
}

func (s *UnstableService) Call() error {
	s.mu.Lock()
	healthy := s.isHealthy
	s.mu.Unlock()

	time.Sleep(100 * time.Millisecond) // 처리 시간 시뮬레이션

	if !healthy {
		return errors.New("서비스 장애 발생")
	}
	return nil
}

func (s *UnstableService) SetHealth(healthy bool) {
	s.mu.Lock()
	s.isHealthy = healthy
	s.mu.Unlock()

	status := "정상"
	if !healthy {
		status = "장애"
	}
	log.Printf("🔧 [Service] 상태 변경: %s", status)
}

func main() {
	log.Println("🚀 Circuit Breaker 시스템 시작\n")

	// Circuit Breaker 생성 (실패 3회 시 Open, 5초 후 Half-Open)
	cb := NewCircuitBreaker("API-Service", 3, 5*time.Second)
	service := &UnstableService{isHealthy: true}

	// 시나리오 1: 정상 동작
	log.Println("=== 시나리오 1: 정상 동작 ===")
	for i := 1; i <= 3; i++ {
		err := cb.Call(func() error {
			return service.Call()
		})
		if err != nil {
			log.Printf("❌ 요청 %d 실패: %v", i, err)
		} else {
			log.Printf("✅ 요청 %d 성공", i)
		}
		time.Sleep(200 * time.Millisecond)
	}

	// 시나리오 2: 서비스 장애 → Circuit OPEN
	log.Println("\n=== 시나리오 2: 서비스 장애 발생 ===")
	service.SetHealth(false)

	for i := 1; i <= 5; i++ {
		err := cb.Call(func() error {
			return service.Call()
		})
		if err == ErrCircuitOpen {
			log.Printf("🚫 요청 %d: Circuit이 OPEN 상태라 즉시 차단됨", i)
		} else if err != nil {
			log.Printf("❌ 요청 %d 실패: %v", i, err)
		} else {
			log.Printf("✅ 요청 %d 성공", i)
		}
		time.Sleep(200 * time.Millisecond)
	}

	state, failures, successes, requests := cb.GetStats()
	log.Printf("\n📊 현재 상태: %s (실패: %d, 성공: %d, 총 요청: %d)\n",
		state, failures, successes, requests)

	// 시나리오 3: 타임아웃 후 Half-Open 전환
	log.Println("\n=== 시나리오 3: Half-Open 전환 대기 ===")
	log.Println("⏳ 5초 대기 중...")
	time.Sleep(5 * time.Second)

	// 서비스는 여전히 장애 상태
	log.Println("\n서비스가 아직 복구되지 않은 상태에서 요청 시도...")
	err := cb.Call(func() error {
		return service.Call()
	})
	if err != nil {
		log.Printf("❌ 복구 확인 실패: %v", err)
	}

	time.Sleep(1 * time.Second)

	// 시나리오 4: 서비스 복구 → Circuit CLOSED
	log.Println("\n=== 시나리오 4: 서비스 복구 ===")
	log.Println("⏳ 5초 더 대기 (Half-Open 재진입)...")
	time.Sleep(5 * time.Second)

	service.SetHealth(true)
	log.Println("서비스가 정상화된 상태에서 요청 시도...\n")

	for i := 1; i <= 3; i++ {
		err := cb.Call(func() error {
			return service.Call()
		})
		if err == ErrCircuitOpen {
			log.Printf("🚫 요청 %d: Circuit이 OPEN 상태", i)
		} else if err != nil {
			log.Printf("❌ 요청 %d 실패: %v", i, err)
		} else {
			log.Printf("✅ 요청 %d 성공", i)
		}
		time.Sleep(200 * time.Millisecond)
	}

	state, failures, successes, requests = cb.GetStats()
	log.Printf("\n📊 최종 상태: %s (실패: %d, 성공: %d, 총 요청: %d)\n",
		state, failures, successes, requests)

	log.Println("\n✨ Circuit Breaker 시뮬레이션 완료!")
	log.Println("💡 학습 포인트:")
	log.Println("   1. 장애 발생 시 Circuit이 OPEN되어 추가 요청 차단")
	log.Println("   2. 타임아웃 후 Half-Open으로 복구 시도")
	log.Println("   3. 서비스 정상화 확인 후 Circuit CLOSED")
	log.Println("   4. 연쇄 장애 방지 및 시스템 안정성 향상")
}
