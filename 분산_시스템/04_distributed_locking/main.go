package main

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

var (
	ErrLockNotAcquired = errors.New("락을 획득하지 못했습니다")
	ErrLockNotHeld     = errors.New("락을 보유하고 있지 않습니다")
)

// Lock 정보
type LockInfo struct {
	owner     string
	acquiredAt time.Time
	ttl       time.Duration
}

// DistributedLock 매니저
type DistributedLockManager struct {
	locks map[string]*LockInfo
	mu    sync.Mutex
}

func NewDistributedLockManager() *DistributedLockManager {
	dlm := &DistributedLockManager{
		locks: make(map[string]*LockInfo),
	}
	// 주기적으로 만료된 락 정리
	go dlm.cleanupExpiredLocks()
	return dlm
}

// 락 획득 시도
func (dlm *DistributedLockManager) TryLock(resource string, owner string, ttl time.Duration) error {
	dlm.mu.Lock()
	defer dlm.mu.Unlock()

	// 이미 락이 존재하는지 확인
	if lock, exists := dlm.locks[resource]; exists {
		// TTL이 만료되지 않았으면 획득 실패
		if time.Since(lock.acquiredAt) < lock.ttl {
			return ErrLockNotAcquired
		}
		// TTL이 만료되었으면 락 정보 삭제
		log.Printf("🕐 [Lock Manager] 만료된 락 삭제: %s (이전 소유자: %s)", resource, lock.owner)
		delete(dlm.locks, resource)
	}

	// 새 락 생성
	dlm.locks[resource] = &LockInfo{
		owner:      owner,
		acquiredAt: time.Now(),
		ttl:        ttl,
	}

	log.Printf("🔒 [Lock Manager] 락 획득: %s → %s (TTL: %s)", resource, owner, ttl)
	return nil
}

// 블로킹 방식으로 락 획득 (재시도 포함)
func (dlm *DistributedLockManager) Lock(resource string, owner string, ttl time.Duration, timeout time.Duration) error {
	start := time.Now()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		err := dlm.TryLock(resource, owner, ttl)
		if err == nil {
			return nil
		}

		if time.Since(start) > timeout {
			return fmt.Errorf("타임아웃: %w", ErrLockNotAcquired)
		}

		<-ticker.C
	}
}

// 락 해제
func (dlm *DistributedLockManager) Unlock(resource string, owner string) error {
	dlm.mu.Lock()
	defer dlm.mu.Unlock()

	lock, exists := dlm.locks[resource]
	if !exists {
		return ErrLockNotHeld
	}

	// 소유자 확인
	if lock.owner != owner {
		return fmt.Errorf("다른 소유자의 락입니다: %s", lock.owner)
	}

	delete(dlm.locks, resource)
	log.Printf("🔓 [Lock Manager] 락 해제: %s ← %s", resource, owner)
	return nil
}

// 만료된 락 정리
func (dlm *DistributedLockManager) cleanupExpiredLocks() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		dlm.mu.Lock()
		for resource, lock := range dlm.locks {
			if time.Since(lock.acquiredAt) > lock.ttl {
				log.Printf("🧹 [Cleanup] 만료된 락 정리: %s (소유자: %s)", resource, lock.owner)
				delete(dlm.locks, resource)
			}
		}
		dlm.mu.Unlock()
	}
}

// 상태 출력
func (dlm *DistributedLockManager) PrintStatus() {
	dlm.mu.Lock()
	defer dlm.mu.Unlock()

	log.Println("\n📊 현재 락 상태:")
	if len(dlm.locks) == 0 {
		log.Println("   (락 없음)")
	} else {
		for resource, lock := range dlm.locks {
			elapsed := time.Since(lock.acquiredAt)
			remaining := lock.ttl - elapsed
			log.Printf("   - %s: 소유자=%s, 남은시간=%.1fs", resource, lock.owner, remaining.Seconds())
		}
	}
}

// 작업 시뮬레이션
func worker(dlm *DistributedLockManager, workerID int, resource string, wg *sync.WaitGroup) {
	defer wg.Done()

	ownerID := fmt.Sprintf("Worker-%d", workerID)

	// 락 획득 시도
	log.Printf("⏳ [%s] 락 획득 시도 중...", ownerID)
	err := dlm.Lock(resource, ownerID, 2*time.Second, 5*time.Second)
	if err != nil {
		log.Printf("❌ [%s] 락 획득 실패: %v", ownerID, err)
		return
	}

	log.Printf("✅ [%s] 락 획득 성공! 작업 시작", ownerID)

	// 작업 수행 (시뮬레이션)
	time.Sleep(time.Duration(500+workerID*100) * time.Millisecond)

	log.Printf("✅ [%s] 작업 완료, 락 해제", ownerID)

	// 락 해제
	if err := dlm.Unlock(resource, ownerID); err != nil {
		log.Printf("❌ [%s] 락 해제 실패: %v", ownerID, err)
	}
}

func main() {
	log.Println("🚀 분산 락 시스템 시작\n")

	dlm := NewDistributedLockManager()
	resource := "shared-database"

	// 시나리오 1: 순차적 락 획득/해제
	log.Println("=== 시나리오 1: 순차적 락 획득/해제 ===")
	dlm.Lock(resource, "Process-A", 3*time.Second, 5*time.Second)
	dlm.PrintStatus()

	time.Sleep(1 * time.Second)
	dlm.Unlock(resource, "Process-A")
	dlm.PrintStatus()

	// 시나리오 2: 동시 접근 (Race Condition 방지)
	log.Println("\n=== 시나리오 2: 5개 Worker의 동시 접근 ===")
	log.Println("💡 락을 통해 한 번에 하나의 Worker만 작업을 수행합니다\n")

	var wg sync.WaitGroup
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go worker(dlm, i, resource, &wg)
		time.Sleep(100 * time.Millisecond) // 약간의 지연
	}

	wg.Wait()
	log.Println("\n✅ 모든 Worker 작업 완료")

	// 시나리오 3: TTL 자동 만료
	log.Println("\n=== 시나리오 3: TTL 자동 만료 테스트 ===")
	dlm.Lock(resource, "Process-B", 2*time.Second, 5*time.Second)
	dlm.PrintStatus()

	log.Println("⏳ 3초 대기 중 (TTL 만료 대기)...")
	time.Sleep(3 * time.Second)

	// TTL이 만료되어 다른 프로세스가 락을 획득할 수 있음
	err := dlm.TryLock(resource, "Process-C", 2*time.Second)
	if err == nil {
		log.Println("✅ TTL 만료 후 새로운 프로세스가 락을 획득했습니다")
		dlm.PrintStatus()
		dlm.Unlock(resource, "Process-C")
	}

	log.Println("\n✨ 분산 락 시뮬레이션 완료!")
	log.Println("💡 학습 포인트:")
	log.Println("   1. 락을 통해 여러 프로세스의 동시 접근을 제어합니다")
	log.Println("   2. TTL을 통해 데드락을 방지합니다")
	log.Println("   3. 실제 환경에서는 Redis의 SETNX를 사용합니다")
}
