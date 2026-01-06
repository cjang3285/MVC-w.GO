package main

import (
	"container/list"
	"fmt"
	"log"
	"sync"
	"time"
)

// CacheItem 캐시 아이템
type CacheItem struct {
	Key       string
	Value     interface{}
	ExpiresAt time.Time
}

// LRUCache LRU 기반 캐시
type LRUCache struct {
	capacity  int
	cache     map[string]*list.Element
	lruList   *list.List
	mu        sync.Mutex
	hits      int64
	misses    int64
	evictions int64
}

func NewLRUCache(capacity int) *LRUCache {
	c := &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		lruList:  list.New(),
	}
	// 주기적으로 만료된 항목 제거
	go c.cleanupExpired()
	return c
}

// Get: 캐시에서 값 가져오기
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, exists := c.cache[key]
	if !exists {
		c.misses++
		return nil, false
	}

	item := elem.Value.(*CacheItem)

	// TTL 확인
	if time.Now().After(item.ExpiresAt) {
		c.lruList.Remove(elem)
		delete(c.cache, key)
		c.misses++
		log.Printf("⏰ [Cache] TTL 만료: %s", key)
		return nil, false
	}

	// LRU: 최근 사용된 항목을 앞으로 이동
	c.lruList.MoveToFront(elem)
	c.hits++

	return item.Value, true
}

// Set: 캐시에 값 저장
func (c *LRUCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiresAt := time.Now().Add(ttl)

	// 이미 존재하는 키면 업데이트
	if elem, exists := c.cache[key]; exists {
		c.lruList.MoveToFront(elem)
		elem.Value = &CacheItem{
			Key:       key,
			Value:     value,
			ExpiresAt: expiresAt,
		}
		log.Printf("🔄 [Cache] 업데이트: %s (TTL: %s)", key, ttl)
		return
	}

	// 용량 초과 시 LRU 항목 제거
	if c.lruList.Len() >= c.capacity {
		c.evictLRU()
	}

	// 새 항목 추가
	item := &CacheItem{
		Key:       key,
		Value:     value,
		ExpiresAt: expiresAt,
	}
	elem := c.lruList.PushFront(item)
	c.cache[key] = elem

	log.Printf("✅ [Cache] 저장: %s = %v (TTL: %s)", key, value, ttl)
}

// Delete: 캐시에서 삭제
func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.cache[key]; exists {
		c.lruList.Remove(elem)
		delete(c.cache, key)
		log.Printf("🗑️  [Cache] 삭제: %s", key)
	}
}

// evictLRU: 가장 오래된 항목 제거
func (c *LRUCache) evictLRU() {
	elem := c.lruList.Back()
	if elem != nil {
		item := elem.Value.(*CacheItem)
		c.lruList.Remove(elem)
		delete(c.cache, item.Key)
		c.evictions++
		log.Printf("🚮 [Cache] LRU 제거: %s (용량 초과)", item.Key)
	}
}

// cleanupExpired: 만료된 항목 정리
func (c *LRUCache) cleanupExpired() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		toRemove := make([]string, 0)

		for key, elem := range c.cache {
			item := elem.Value.(*CacheItem)
			if now.After(item.ExpiresAt) {
				toRemove = append(toRemove, key)
			}
		}

		for _, key := range toRemove {
			if elem, exists := c.cache[key]; exists {
				c.lruList.Remove(elem)
				delete(c.cache, key)
				log.Printf("🧹 [Cleanup] 만료된 항목 제거: %s", key)
			}
		}
		c.mu.Unlock()
	}
}

// GetStats: 통계 정보
func (c *LRUCache) GetStats() (hits, misses, evictions int64, hitRate float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	total := c.hits + c.misses
	hitRate = 0
	if total > 0 {
		hitRate = float64(c.hits) / float64(total) * 100
	}

	return c.hits, c.misses, c.evictions, hitRate
}

// 데이터베이스 시뮬레이션
func fetchFromDB(key string) (string, error) {
	log.Printf("💾 [DB] 조회 중: %s (느림)", key)
	time.Sleep(500 * time.Millisecond) // DB 조회는 느림
	return fmt.Sprintf("Data for %s from DB", key), nil
}

// Cache-Aside 패턴
func getDataWithCache(cache *LRUCache, key string) (string, error) {
	// 1. 캐시에서 먼저 조회
	if value, found := cache.Get(key); found {
		log.Printf("🎯 [Cache HIT] %s", key)
		return value.(string), nil
	}

	// 2. 캐시 미스: DB에서 조회
	log.Printf("❌ [Cache MISS] %s", key)
	value, err := fetchFromDB(key)
	if err != nil {
		return "", err
	}

	// 3. 캐시에 저장
	cache.Set(key, value, 10*time.Second)

	return value, nil
}

func main() {
	log.Println("🚀 분산 캐싱 시스템 시작\n")

	// 캐시 생성 (용량: 5개)
	cache := NewLRUCache(5)

	// 시나리오 1: Cache-Aside 패턴
	log.Println("=== 시나리오 1: Cache-Aside 패턴 ===")
	log.Println("💡 첫 조회는 DB에서, 두 번째 조회는 캐시에서 가져옵니다\n")

	// 첫 번째 조회 (Cache Miss)
	data, _ := getDataWithCache(cache, "user:1")
	log.Printf("📄 결과: %s\n", data)

	time.Sleep(500 * time.Millisecond)

	// 두 번째 조회 (Cache Hit)
	data, _ = getDataWithCache(cache, "user:1")
	log.Printf("📄 결과: %s\n", data)

	// 시나리오 2: LRU 제거
	log.Println("\n=== 시나리오 2: LRU 제거 (용량 초과) ===")
	log.Println("💡 용량이 5개이므로 6번째 항목 추가 시 가장 오래된 항목이 제거됩니다\n")

	cache.Set("item:1", "Data 1", 30*time.Second)
	cache.Set("item:2", "Data 2", 30*time.Second)
	cache.Set("item:3", "Data 3", 30*time.Second)
	cache.Set("item:4", "Data 4", 30*time.Second)
	cache.Set("item:5", "Data 5", 30*time.Second)

	log.Println("\n현재 캐시 상태:")
	for i := 1; i <= 5; i++ {
		key := fmt.Sprintf("item:%d", i)
		if _, found := cache.Get(key); found {
			log.Printf("   ✅ %s 존재", key)
		}
	}

	log.Println("\n6번째 항목 추가 (LRU 제거 발생):")
	cache.Set("item:6", "Data 6", 30*time.Second)

	log.Println("\n최종 캐시 상태:")
	for i := 1; i <= 6; i++ {
		key := fmt.Sprintf("item:%d", i)
		if _, found := cache.Get(key); found {
			log.Printf("   ✅ %s 존재", key)
		} else {
			log.Printf("   ❌ %s 제거됨", key)
		}
	}

	// 시나리오 3: TTL 만료
	log.Println("\n=== 시나리오 3: TTL 만료 ===")
	cache.Set("temp:1", "Temporary Data", 3*time.Second)

	log.Println("⏳ 즉시 조회:")
	if val, found := cache.Get("temp:1"); found {
		log.Printf("   ✅ 조회 성공: %v", val)
	}

	log.Println("\n⏳ 4초 대기 중...")
	time.Sleep(4 * time.Second)

	log.Println("만료 후 조회:")
	if _, found := cache.Get("temp:1"); !found {
		log.Println("   ❌ TTL 만료로 조회 실패")
	}

	// 시나리오 4: 성능 비교
	log.Println("\n=== 시나리오 4: 캐시 vs DB 성능 비교 ===")

	// DB 직접 조회 (10회)
	log.Println("📊 DB 직접 조회 (10회):")
	startDB := time.Now()
	for i := 1; i <= 10; i++ {
		fetchFromDB(fmt.Sprintf("product:%d", i))
	}
	dbTime := time.Since(startDB)
	log.Printf("   총 시간: %s\n", dbTime)

	// 캐시 사용 (10회)
	log.Println("📊 캐시 사용 (10회):")
	startCache := time.Now()
	for i := 1; i <= 10; i++ {
		getDataWithCache(cache, fmt.Sprintf("product:%d", i%3)) // 3개 항목 반복
	}
	cacheTime := time.Since(startCache)
	log.Printf("   총 시간: %s", cacheTime)
	log.Printf("   🚀 성능 향상: %.1fx 빠름\n", float64(dbTime)/float64(cacheTime))

	// 통계 출력
	hits, misses, evictions, hitRate := cache.GetStats()
	log.Println("\n📊 캐시 통계:")
	log.Printf("   Hits: %d", hits)
	log.Printf("   Misses: %d", misses)
	log.Printf("   Evictions: %d", evictions)
	log.Printf("   Hit Rate: %.1f%%", hitRate)

	log.Println("\n✨ 분산 캐싱 시뮬레이션 완료!")
	log.Println("💡 학습 포인트:")
	log.Println("   1. 캐시를 통한 성능 향상")
	log.Println("   2. LRU 알고리즘으로 메모리 관리")
	log.Println("   3. TTL을 통한 자동 만료")
	log.Println("   4. Cache Hit Rate가 높을수록 효과적")
	log.Println("   5. 실제 환경: Redis, Memcached 사용")
}
