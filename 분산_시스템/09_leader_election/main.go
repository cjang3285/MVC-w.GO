package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

type NodeStatus int

const (
	StatusFollower NodeStatus = iota
	StatusCandidate
	StatusLeader
	StatusDown
)

func (s NodeStatus) String() string {
	switch s {
	case StatusFollower:
		return "Follower"
	case StatusCandidate:
		return "Candidate"
	case StatusLeader:
		return "Leader"
	case StatusDown:
		return "Down"
	default:
		return "Unknown"
	}
}

type Node struct {
	id            int
	status        NodeStatus
	currentLeader int
	mu            sync.Mutex
	peers         []*Node
	lastHeartbeat time.Time
	isAlive       bool
}

func NewNode(id int) *Node {
	return &Node{
		id:            id,
		status:        StatusFollower,
		currentLeader: -1,
		lastHeartbeat: time.Now(),
		isAlive:       true,
	}
}

// StartElection: 선거 시작 (Bully Algorithm)
func (n *Node) StartElection() {
	n.mu.Lock()
	if !n.isAlive {
		n.mu.Unlock()
		return
	}
	n.status = StatusCandidate
	myID := n.id
	n.mu.Unlock()

	log.Printf("🗳️  [Node %d] 선거 시작!", myID)

	// 자신보다 높은 ID를 가진 노드들에게 선거 메시지 전송
	higherNodes := make([]*Node, 0)
	for _, peer := range n.peers {
		peer.mu.Lock()
		if peer.id > myID && peer.isAlive {
			higherNodes = append(higherNodes, peer)
		}
		peer.mu.Unlock()
	}

	// 더 높은 ID의 노드가 없으면 자신이 리더
	if len(higherNodes) == 0 {
		n.becomeLeader()
		return
	}

	// 더 높은 ID의 노드들에게 선거 메시지 전송
	responses := 0
	for _, higher := range higherNodes {
		if higher.respondToElection(myID) {
			responses++
		}
	}

	// 응답이 있으면 양보 (더 높은 ID의 노드가 리더가 됨)
	if responses > 0 {
		log.Printf("   [Node %d] 더 높은 ID의 노드가 있어 양보합니다", myID)
		n.mu.Lock()
		n.status = StatusFollower
		n.mu.Unlock()
	} else {
		// 응답이 없으면 자신이 리더
		n.becomeLeader()
	}
}

func (n *Node) respondToElection(senderID int) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.isAlive {
		return false
	}

	// 자신이 더 높은 ID를 가지면 응답하고 스스로 선거 시작
	if n.id > senderID {
		log.Printf("   [Node %d] Node %d의 선거에 응답 (내가 더 높은 ID)", n.id, senderID)
		go n.StartElection()
		return true
	}

	return false
}

func (n *Node) becomeLeader() {
	n.mu.Lock()
	n.status = StatusLeader
	n.currentLeader = n.id
	myID := n.id
	n.mu.Unlock()

	log.Printf("👑 [Node %d] 리더로 선출되었습니다!", myID)

	// 모든 노드에게 리더 알림
	n.announceLeader()

	// 하트비트 시작
	go n.sendHeartbeats()
}

func (n *Node) announceLeader() {
	n.mu.Lock()
	myID := n.id
	n.mu.Unlock()

	for _, peer := range n.peers {
		peer.receiveLeaderAnnouncement(myID)
	}
}

func (n *Node) receiveLeaderAnnouncement(leaderID int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.isAlive {
		return
	}

	n.currentLeader = leaderID
	n.lastHeartbeat = time.Now()
	if n.status != StatusLeader {
		n.status = StatusFollower
	}
}

func (n *Node) sendHeartbeats() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		n.mu.Lock()
		if n.status != StatusLeader || !n.isAlive {
			n.mu.Unlock()
			return
		}
		myID := n.id
		n.mu.Unlock()

		// 모든 팔로워에게 하트비트 전송
		for _, peer := range n.peers {
			peer.receiveHeartbeat(myID)
		}
	}
}

func (n *Node) receiveHeartbeat(leaderID int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.isAlive {
		return
	}

	n.currentLeader = leaderID
	n.lastHeartbeat = time.Now()
}

func (n *Node) monitorLeader() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		n.mu.Lock()
		if !n.isAlive {
			n.mu.Unlock()
			return
		}

		// 리더가 아닌 경우에만 체크
		if n.status != StatusLeader {
			elapsed := time.Since(n.lastHeartbeat)
			if elapsed > 3*time.Second {
				currentLeader := n.currentLeader
				n.mu.Unlock()

				log.Printf("💔 [Node %d] 리더(Node %d)의 하트비트 감지 안됨 (%.1fs)",
					n.id, currentLeader, elapsed.Seconds())
				n.StartElection()
				continue
			}
		}
		n.mu.Unlock()
	}
}

func (n *Node) crash() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.isAlive = false
	n.status = StatusDown
	log.Printf("💥 [Node %d] 장애 발생!", n.id)
}

func (n *Node) recover() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.isAlive = true
	n.status = StatusFollower
	n.lastHeartbeat = time.Now()
	log.Printf("🔧 [Node %d] 복구됨", n.id)
}

func printClusterStatus(nodes []*Node) {
	log.Println("\n" + "=".repeat(50))
	log.Println("📊 클러스터 상태:")
	for _, node := range nodes {
		node.mu.Lock()
		status := "💚"
		if !node.isAlive {
			status = "💔"
		} else if node.status == StatusLeader {
			status = "👑"
		}
		log.Printf("   %s Node %d: %s (리더: Node %d)",
			status, node.id, node.status, node.currentLeader)
		node.mu.Unlock()
	}
	log.Println("=".repeat(50))
}

func main() {
	rand.Seed(time.Now().UnixNano())
	log.Println("🚀 리더 선출 시스템 시작 (Bully Algorithm)\n")

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

	// 시나리오 1: 초기 리더 선출
	log.Println("=== 시나리오 1: 초기 리더 선출 ===\n")

	// 모든 노드 시작
	for _, node := range nodes {
		go node.monitorLeader()
	}

	// 랜덤 노드가 선거 시작
	starter := nodes[rand.Intn(len(nodes))]
	log.Printf("🎲 Node %d가 선거를 시작합니다\n", starter.id)
	starter.StartElection()

	time.Sleep(2 * time.Second)
	printClusterStatus(nodes)

	// 시나리오 2: 리더 장애
	log.Println("\n=== 시나리오 2: 리더 장애 및 재선출 ===\n")

	// 현재 리더 찾기
	var currentLeader *Node
	for _, node := range nodes {
		node.mu.Lock()
		if node.status == StatusLeader {
			currentLeader = node
		}
		node.mu.Unlock()
	}

	if currentLeader != nil {
		log.Printf("💥 현재 리더(Node %d)에 장애 발생\n", currentLeader.id)
		currentLeader.crash()

		log.Println("⏳ 새로운 리더 선출 대기 중...")
		time.Sleep(4 * time.Second)

		printClusterStatus(nodes)
	}

	// 시나리오 3: 리더 복구 (더 높은 ID)
	log.Println("\n=== 시나리오 3: 장애 노드 복구 ===\n")

	if currentLeader != nil {
		log.Printf("🔧 Node %d 복구\n", currentLeader.id)
		currentLeader.recover()
		go currentLeader.monitorLeader()

		// 복구된 노드가 더 높은 ID를 가지면 선거 시작
		currentLeader.mu.Lock()
		recoveredID := currentLeader.id
		currentLeader.mu.Unlock()

		// 현재 리더 ID 확인
		var activeLeaderID int
		for _, node := range nodes {
			node.mu.Lock()
			if node.status == StatusLeader && node.isAlive {
				activeLeaderID = node.id
			}
			node.mu.Unlock()
		}

		if recoveredID > activeLeaderID {
			log.Printf("💡 복구된 Node %d가 현재 리더보다 높은 ID를 가짐 → 선거 시작\n", recoveredID)
			currentLeader.StartElection()
		}

		time.Sleep(2 * time.Second)
		printClusterStatus(nodes)
	}

	// 시나리오 4: 다중 노드 장애
	log.Println("\n=== 시나리오 4: 다중 노드 장애 ===\n")

	log.Println("💥 3개 노드에 동시 장애 발생")
	for i := 0; i < 3; i++ {
		nodes[i].crash()
	}

	time.Sleep(4 * time.Second)
	printClusterStatus(nodes)

	log.Println("\n✨ 리더 선출 시뮬레이션 완료!")
	log.Println("💡 학습 포인트:")
	log.Println("   1. Bully Algorithm: 가장 높은 ID가 리더")
	log.Println("   2. 리더 장애 시 자동 재선출")
	log.Println("   3. 하트비트를 통한 장애 감지")
	log.Println("   4. 실제 환경: ZooKeeper, etcd의 리더 선출 사용")
}

// 문자열 반복 헬퍼
type stringRepeater string

func (s stringRepeater) repeat(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += string(s)
	}
	return result
}
