package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

// ServiceInstance 서비스 인스턴스 정보
type ServiceInstance struct {
	ID       string
	Name     string
	Host     string
	Port     int
	Metadata map[string]string
	Status   string // "healthy", "unhealthy"
	LastSeen time.Time
}

// ServiceRegistry 서비스 레지스트리
type ServiceRegistry struct {
	services map[string][]*ServiceInstance // key: 서비스 이름, value: 인스턴스 목록
	mu       sync.RWMutex
}

func NewServiceRegistry() *ServiceRegistry {
	sr := &ServiceRegistry{
		services: make(map[string][]*ServiceInstance),
	}
	// Health Check 시작
	go sr.healthCheckLoop()
	return sr
}

// Register: 서비스 등록
func (sr *ServiceRegistry) Register(instance *ServiceInstance) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	instance.Status = "healthy"
	instance.LastSeen = time.Now()

	if _, exists := sr.services[instance.Name]; !exists {
		sr.services[instance.Name] = make([]*ServiceInstance, 0)
	}

	sr.services[instance.Name] = append(sr.services[instance.Name], instance)
	log.Printf("✅ [Registry] 서비스 등록: %s (ID: %s, %s:%d)",
		instance.Name, instance.ID, instance.Host, instance.Port)
}

// Deregister: 서비스 해제
func (sr *ServiceRegistry) Deregister(serviceName string, instanceID string) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	instances, exists := sr.services[serviceName]
	if !exists {
		return
	}

	for i, inst := range instances {
		if inst.ID == instanceID {
			sr.services[serviceName] = append(instances[:i], instances[i+1:]...)
			log.Printf("❌ [Registry] 서비스 해제: %s (ID: %s)", serviceName, instanceID)
			return
		}
	}
}

// Discover: 서비스 조회 (건강한 인스턴스만)
func (sr *ServiceRegistry) Discover(serviceName string) []*ServiceInstance {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	instances, exists := sr.services[serviceName]
	if !exists {
		return []*ServiceInstance{}
	}

	// 건강한 인스턴스만 반환
	healthy := make([]*ServiceInstance, 0)
	for _, inst := range instances {
		if inst.Status == "healthy" {
			healthy = append(healthy, inst)
		}
	}

	return healthy
}

// GetInstance: 랜덤하게 하나의 인스턴스 선택
func (sr *ServiceRegistry) GetInstance(serviceName string) *ServiceInstance {
	instances := sr.Discover(serviceName)
	if len(instances) == 0 {
		return nil
	}
	return instances[rand.Intn(len(instances))]
}

// Heartbeat: 서비스가 살아있음을 알림
func (sr *ServiceRegistry) Heartbeat(serviceName string, instanceID string) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	instances, exists := sr.services[serviceName]
	if !exists {
		return
	}

	for _, inst := range instances {
		if inst.ID == instanceID {
			inst.LastSeen = time.Now()
			if inst.Status != "healthy" {
				inst.Status = "healthy"
				log.Printf("💚 [Registry] 서비스 복구: %s (ID: %s)", serviceName, instanceID)
			}
			return
		}
	}
}

// Health Check Loop
func (sr *ServiceRegistry) healthCheckLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		sr.mu.Lock()
		for serviceName, instances := range sr.services {
			for _, inst := range instances {
				elapsed := time.Since(inst.LastSeen)
				if elapsed > 5*time.Second && inst.Status == "healthy" {
					inst.Status = "unhealthy"
					log.Printf("💔 [Health Check] 서비스 비정상: %s (ID: %s) - 마지막 하트비트: %.1fs 전",
						serviceName, inst.ID, elapsed.Seconds())
				}
			}
		}
		sr.mu.Unlock()
	}
}

// PrintStatus: 현재 등록된 서비스 출력
func (sr *ServiceRegistry) PrintStatus() {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	log.Println("\n📊 서비스 레지스트리 상태:")
	if len(sr.services) == 0 {
		log.Println("   (등록된 서비스 없음)")
		return
	}

	for serviceName, instances := range sr.services {
		log.Printf("   🔹 %s (%d개 인스턴스):", serviceName, len(instances))
		for _, inst := range instances {
			status := "💚"
			if inst.Status == "unhealthy" {
				status = "💔"
			}
			log.Printf("      %s ID: %s, Host: %s:%d, Status: %s",
				status, inst.ID, inst.Host, inst.Port, inst.Status)
		}
	}
	log.Println()
}

// 서비스 시뮬레이션
func simulateService(registry *ServiceRegistry, serviceName string, instanceID string, duration time.Duration) {
	host := "localhost"
	port := 8000 + rand.Intn(1000)

	instance := &ServiceInstance{
		ID:   instanceID,
		Name: serviceName,
		Host: host,
		Port: port,
		Metadata: map[string]string{
			"version": "1.0.0",
			"region":  "kr-central",
		},
	}

	// 서비스 등록
	registry.Register(instance)

	// 주기적으로 Heartbeat 전송
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeout := time.After(duration)
	for {
		select {
		case <-ticker.C:
			registry.Heartbeat(serviceName, instanceID)
		case <-timeout:
			// 서비스 종료
			registry.Deregister(serviceName, instanceID)
			return
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	log.Println("🚀 서비스 디스커버리 시스템 시작\n")

	registry := NewServiceRegistry()

	// 시나리오 1: 서비스 등록 및 조회
	log.Println("=== 시나리오 1: 서비스 등록 및 조회 ===")

	// 여러 서비스 인스턴스 등록
	go simulateService(registry, "user-service", "user-1", 10*time.Second)
	go simulateService(registry, "user-service", "user-2", 10*time.Second)
	go simulateService(registry, "order-service", "order-1", 10*time.Second)

	time.Sleep(1 * time.Second)
	registry.PrintStatus()

	// 서비스 조회
	log.Println("=== 서비스 조회 ===")
	userServices := registry.Discover("user-service")
	log.Printf("🔍 user-service 인스턴스 %d개 발견:", len(userServices))
	for _, inst := range userServices {
		log.Printf("   - %s:%d (ID: %s)", inst.Host, inst.Port, inst.ID)
	}

	// 랜덤 인스턴스 선택
	selectedInstance := registry.GetInstance("user-service")
	if selectedInstance != nil {
		log.Printf("🎯 선택된 인스턴스: %s:%d (ID: %s)\n",
			selectedInstance.Host, selectedInstance.Port, selectedInstance.ID)
	}

	// 시나리오 2: Health Check 및 Auto Scaling
	log.Println("\n=== 시나리오 2: Auto Scaling 시뮬레이션 ===")
	log.Println("💡 부하에 따라 서비스 인스턴스를 동적으로 추가/제거합니다\n")

	// 초기 인스턴스
	go simulateService(registry, "api-service", "api-1", 8*time.Second)
	time.Sleep(1 * time.Second)

	// 부하 증가 - 인스턴스 추가
	log.Println("📈 부하 증가 감지! 인스턴스 추가...")
	go simulateService(registry, "api-service", "api-2", 6*time.Second)
	go simulateService(registry, "api-service", "api-3", 6*time.Second)
	time.Sleep(1 * time.Second)

	registry.PrintStatus()

	// 시나리오 3: 장애 시뮬레이션
	log.Println("\n=== 시나리오 3: 장애 감지 및 복구 ===")
	log.Println("💡 하트비트가 중단되면 자동으로 unhealthy로 표시됩니다\n")

	// 장애 발생할 서비스 (하트비트 중단)
	registry.Register(&ServiceInstance{
		ID:       "faulty-1",
		Name:     "payment-service",
		Host:     "localhost",
		Port:     9000,
		Metadata: map[string]string{"version": "1.0.0"},
	})

	log.Println("⏳ 6초 대기 (Health Check 모니터링)...")
	for i := 1; i <= 6; i++ {
		time.Sleep(1 * time.Second)
		if i == 3 {
			registry.PrintStatus()
		}
	}

	registry.PrintStatus()

	// 서비스 복구
	log.Println("🔧 서비스 복구 중...")
	registry.Heartbeat("payment-service", "faulty-1")
	time.Sleep(500 * time.Millisecond)
	registry.PrintStatus()

	log.Println("\n✨ 서비스 디스커버리 시뮬레이션 완료!")
	log.Println("💡 학습 포인트:")
	log.Println("   1. 동적 서비스 등록 및 해제")
	log.Println("   2. Health Check를 통한 장애 감지")
	log.Println("   3. Auto Scaling 지원")
	log.Println("   4. 실제 환경: Consul, etcd, Kubernetes Service Discovery")
}
