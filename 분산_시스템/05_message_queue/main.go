package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Message 구조체
type Message struct {
	ID        string
	Content   string
	Timestamp time.Time
	Retries   int
}

// MessageQueue 구조체
type MessageQueue struct {
	queue    []*Message
	mu       sync.Mutex
	notEmpty *sync.Cond
	maxSize  int
}

func NewMessageQueue(maxSize int) *MessageQueue {
	mq := &MessageQueue{
		queue:   make([]*Message, 0),
		maxSize: maxSize,
	}
	mq.notEmpty = sync.NewCond(&mq.mu)
	return mq
}

// Enqueue: 메시지를 큐에 추가 (Producer)
func (mq *MessageQueue) Enqueue(msg *Message) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if len(mq.queue) >= mq.maxSize {
		return fmt.Errorf("큐가 가득 찼습니다 (최대: %d)", mq.maxSize)
	}

	mq.queue = append(mq.queue, msg)
	log.Printf("📥 [Queue] 메시지 추가: ID=%s, Content='%s' (큐 크기: %d)",
		msg.ID, msg.Content, len(mq.queue))

	// Consumer에게 메시지가 있음을 알림
	mq.notEmpty.Signal()
	return nil
}

// Dequeue: 큐에서 메시지 꺼내기 (Consumer)
func (mq *MessageQueue) Dequeue() *Message {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	// 큐가 비어있으면 대기
	for len(mq.queue) == 0 {
		mq.notEmpty.Wait()
	}

	msg := mq.queue[0]
	mq.queue = mq.queue[1:]

	log.Printf("📤 [Queue] 메시지 꺼냄: ID=%s (큐 크기: %d)", msg.ID, len(mq.queue))
	return msg
}

// Size: 현재 큐 크기
func (mq *MessageQueue) Size() int {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	return len(mq.queue)
}

// Producer: 메시지를 생성하여 큐에 전송
func producer(mq *MessageQueue, producerID int, count int, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 1; i <= count; i++ {
		msg := &Message{
			ID:        fmt.Sprintf("P%d-M%d", producerID, i),
			Content:   fmt.Sprintf("Producer %d의 메시지 #%d", producerID, i),
			Timestamp: time.Now(),
			Retries:   0,
		}

		if err := mq.Enqueue(msg); err != nil {
			log.Printf("❌ [Producer %d] 메시지 전송 실패: %v", producerID, err)
		}

		time.Sleep(time.Duration(100+producerID*50) * time.Millisecond)
	}

	log.Printf("✅ [Producer %d] 모든 메시지 전송 완료", producerID)
}

// Consumer: 큐에서 메시지를 가져와 처리
func consumer(mq *MessageQueue, consumerID int, wg *sync.WaitGroup, stopChan <-chan bool) {
	defer wg.Done()

	for {
		select {
		case <-stopChan:
			log.Printf("🛑 [Consumer %d] 종료", consumerID)
			return
		default:
			// 메시지 가져오기 (타임아웃 포함)
			done := make(chan *Message)
			go func() {
				msg := mq.Dequeue()
				done <- msg
			}()

			select {
			case msg := <-done:
				processMessage(consumerID, msg)
			case <-time.After(2 * time.Second):
				// 타임아웃: 큐가 비어있고 더 이상 메시지가 없음
				return
			}
		}
	}
}

// 메시지 처리 시뮬레이션
func processMessage(consumerID int, msg *Message) {
	log.Printf("⚙️  [Consumer %d] 메시지 처리 중: ID=%s", consumerID, msg.ID)

	// 처리 시간 시뮬레이션
	time.Sleep(time.Duration(200+consumerID*100) * time.Millisecond)

	// 처리 완료
	log.Printf("✅ [Consumer %d] 메시지 처리 완료: ID=%s, Content='%s'",
		consumerID, msg.ID, msg.Content)
}

func main() {
	log.Println("🚀 메시지 큐 시스템 시작\n")

	// 메시지 큐 생성 (최대 100개 메시지)
	mq := NewMessageQueue(100)

	// 시나리오 1: 단일 Producer, 단일 Consumer
	log.Println("=== 시나리오 1: 1 Producer → 1 Consumer ===")
	var wg1 sync.WaitGroup
	stopChan1 := make(chan bool)

	wg1.Add(2)
	go producer(mq, 1, 5, &wg1)
	go consumer(mq, 1, &wg1, stopChan1)

	wg1.Wait()
	log.Printf("\n📊 큐 크기: %d\n", mq.Size())

	// 시나리오 2: 다중 Producer, 다중 Consumer
	log.Println("\n=== 시나리오 2: 3 Producers → 3 Consumers ===")
	log.Println("💡 여러 Consumer가 병렬로 메시지를 처리합니다\n")

	var wg2 sync.WaitGroup
	stopChan2 := make(chan bool)

	// 3개의 Producer 시작
	for i := 1; i <= 3; i++ {
		wg2.Add(1)
		go producer(mq, i, 5, &wg2)
	}

	// 3개의 Consumer 시작
	for i := 1; i <= 3; i++ {
		wg2.Add(1)
		go consumer(mq, i, &wg2, stopChan2)
	}

	wg2.Wait()
	log.Printf("\n📊 최종 큐 크기: %d\n", mq.Size())

	// 시나리오 3: 트래픽 급증 시뮬레이션
	log.Println("\n=== 시나리오 3: 트래픽 급증 (Burst Traffic) ===")
	log.Println("💡 많은 메시지가 빠르게 들어오고, Consumer가 천천히 처리합니다\n")

	var wg3 sync.WaitGroup

	// 빠르게 메시지 생성
	wg3.Add(1)
	go func() {
		defer wg3.Done()
		for i := 1; i <= 20; i++ {
			msg := &Message{
				ID:        fmt.Sprintf("BURST-M%d", i),
				Content:   fmt.Sprintf("급증 메시지 #%d", i),
				Timestamp: time.Now(),
			}
			mq.Enqueue(msg)
			time.Sleep(50 * time.Millisecond)
		}
		log.Println("✅ [Burst Producer] 메시지 전송 완료")
	}()

	// 잠시 대기 후 큐 상태 확인
	time.Sleep(1 * time.Second)
	log.Printf("📊 현재 큐에 쌓인 메시지: %d개\n", mq.Size())

	// Consumer 시작 (천천히 처리)
	stopChan3 := make(chan bool)
	for i := 1; i <= 2; i++ {
		wg3.Add(1)
		go consumer(mq, i, &wg3, stopChan3)
	}

	wg3.Wait()

	log.Println("\n✨ 메시지 큐 시뮬레이션 완료!")
	log.Println("💡 학습 포인트:")
	log.Println("   1. Producer와 Consumer의 비동기 통신")
	log.Println("   2. 메시지 큐를 통한 트래픽 버퍼링")
	log.Println("   3. 여러 Consumer를 통한 병렬 처리")
	log.Println("   4. 실제 환경: RabbitMQ, Kafka, AWS SQS 등 사용")
}
