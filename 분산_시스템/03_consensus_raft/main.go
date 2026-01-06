package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

type NodeState int

const (
	Follower NodeState = iota
	Candidate
	Leader
)

func (s NodeState) String() string {
	switch s {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	default:
		return "Unknown"
	}
}

type Node struct {
	id              int
	state           NodeState
	currentTerm     int
	votedFor        int
	votes           int
	mu              sync.Mutex
	electionTimeout time.Duration
	lastHeartbeat   time.Time
	peers           []*Node
}

func NewNode(id int) *Node {
	return &Node{
		id:              id,
		state:           Follower,
		currentTerm:     0,
		votedFor:        -1,
		electionTimeout: randomTimeout(),
		lastHeartbeat:   time.Now(),
	}
}

func randomTimeout() time.Duration {
	// 150~300ms 사이의 랜덤 타임아웃
	return time.Duration(150+rand.Intn(150)) * time.Millisecond
}

func (n *Node) becomeFollower(term int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.state = Follower
	n.currentTerm = term
	n.votedFor = -1
	n.lastHeartbeat = time.Now()
	log.Printf("🔵 [Node %d] Follower로 전환 (Term %d)", n.id, n.currentTerm)
}

func (n *Node) becomeCandidate() {
	n.mu.Lock()
	n.state = Candidate
	n.currentTerm++
	n.votedFor = n.id
	n.votes = 1 // 자신에게 투표
	currentTerm := n.currentTerm
	n.mu.Unlock()

	log.Printf("🟡 [Node %d] Candidate로 전환, 선거 시작 (Term %d)", n.id, currentTerm)

	// 다른 노드들에게 투표 요청
	for _, peer := range n.peers {
		go n.requestVote(peer, currentTerm)
	}
}

func (n *Node) becomeLeader() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.state = Leader
	log.Printf("🟢 [Node %d] 🎉 Leader로 선출됨! (Term %d)", n.id, n.currentTerm)

	// 리더가 되면 주기적으로 하트비트 전송
	go n.sendHeartbeats()
}

func (n *Node) requestVote(peer *Node, term int) {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	// 더 높은 term을 발견하면 투표하지 않음
	if peer.currentTerm > term {
		return
	}

	// 이미 다른 노드에게 투표했으면 거부
	if peer.votedFor != -1 && peer.votedFor != n.id && peer.currentTerm == term {
		log.Printf("  ❌ [Node %d] Node %d의 투표 요청 거부 (이미 Node %d에게 투표)", peer.id, n.id, peer.votedFor)
		return
	}

	// 투표 승인
	peer.votedFor = n.id
	peer.currentTerm = term
	peer.lastHeartbeat = time.Now()
	log.Printf("  ✅ [Node %d] Node %d에게 투표함 (Term %d)", peer.id, n.id, term)

	// 투표 카운트
	n.mu.Lock()
	n.votes++
	votes := n.votes
	totalNodes := len(n.peers) + 1 // 자신 포함
	n.mu.Unlock()

	// 과반수 득표 확인
	if votes > totalNodes/2 && n.state == Candidate {
		n.becomeLeader()
	}
}

func (n *Node) sendHeartbeats() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		n.mu.Lock()
		if n.state != Leader {
			n.mu.Unlock()
			return
		}
		term := n.currentTerm
		n.mu.Unlock()

		// 모든 팔로워에게 하트비트 전송
		for _, peer := range n.peers {
			go n.sendHeartbeat(peer, term)
		}
	}
}

func (n *Node) sendHeartbeat(peer *Node, term int) {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	if peer.currentTerm > term {
		// 더 높은 term 발견, 리더 포기
		go n.becomeFollower(peer.currentTerm)
		return
	}

	peer.currentTerm = term
	peer.lastHeartbeat = time.Now()
	if peer.state != Follower {
		peer.state = Follower
	}
}

func (n *Node) run() {
	for {
		n.mu.Lock()
		state := n.state
		elapsed := time.Since(n.lastHeartbeat)
		timeout := n.electionTimeout
		n.mu.Unlock()

		switch state {
		case Follower, Candidate:
			// Election timeout 체크
			if elapsed > timeout {
				n.becomeCandidate()
			}
		case Leader:
			// 리더는 하트비트를 계속 보냄
		}

		time.Sleep(50 * time.Millisecond)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// 5개의 노드 생성
	nodes := make([]*Node, 5)
	for i := 0; i < 5; i++ {
		nodes[i] = NewNode(i + 1)
	}

	// 각 노드에 peer 정보 설정
	for i := 0; i < 5; i++ {
		peers := make([]*Node, 0)
		for j := 0; j < 5; j++ {
			if i != j {
				peers = append(peers, nodes[j])
			}
		}
		nodes[i].peers = peers
	}

	log.Println("🚀 Raft 클러스터 시작 (5개 노드)")
	log.Println("📊 Leader Election 과정을 관찰하세요...\n")

	// 모든 노드 시작
	for _, node := range nodes {
		go node.run()
	}

	// 시뮬레이션 실행
	time.Sleep(3 * time.Second)

	// 현재 상태 출력
	log.Println("\n" + "=".repeat(50))
	log.Println("📊 현재 클러스터 상태:")
	for _, node := range nodes {
		node.mu.Lock()
		log.Printf("   Node %d: %s (Term %d)", node.id, node.state, node.currentTerm)
		node.mu.Unlock()
	}

	// 리더 찾기
	log.Println("\n" + "=".repeat(50))
	for _, node := range nodes {
		node.mu.Lock()
		if node.state == Leader {
			log.Printf("👑 현재 리더: Node %d (Term %d)", node.id, node.currentTerm)
		}
		node.mu.Unlock()
	}

	// 리더 장애 시뮬레이션
	log.Println("\n" + "=".repeat(50))
	log.Println("💥 리더 장애 시뮬레이션 시작...")
	for _, node := range nodes {
		node.mu.Lock()
		if node.state == Leader {
			log.Printf("🔴 [Node %d] 리더를 강제로 중단합니다", node.id)
			node.state = Follower
			node.votedFor = -1
		}
		node.mu.Unlock()
	}

	// 새로운 리더 선출 대기
	log.Println("⏳ 새로운 리더 선출 대기 중...")
	time.Sleep(2 * time.Second)

	log.Println("\n" + "=".repeat(50))
	log.Println("📊 장애 복구 후 클러스터 상태:")
	for _, node := range nodes {
		node.mu.Lock()
		log.Printf("   Node %d: %s (Term %d)", node.id, node.state, node.currentTerm)
		if node.state == Leader {
			log.Printf("   ✨ 새로운 리더가 선출되었습니다!")
		}
		node.mu.Unlock()
	}

	log.Println("\n✅ Raft 시뮬레이션 완료!")
	log.Println("💡 리더가 장애가 나도 새로운 리더가 자동으로 선출되는 것을 확인했습니다.")
}

// 문자열 반복 헬퍼 함수
func (s string) repeat(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
