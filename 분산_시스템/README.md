# 분산 시스템 학습 프로젝트

Go 언어로 구현하는 분산 시스템 핵심 개념 10가지 실습 프로젝트입니다.

## 🎯 학습 목표

분산 시스템의 핵심 개념을 이론뿐만 아니라 직접 구현해보면서 깊이 이해하기

## 📚 다루는 핵심 개념

### 1. Client-Server 모델 (01_client_server)
- 기본적인 클라이언트-서버 통신 구현
- TCP/HTTP 기반 통신
- 요청-응답 패턴 이해

### 2. Load Balancing (02_load_balancing)
- Round Robin, Least Connection 알고리즘
- 여러 서버에 트래픽 분산
- Health Check 구현

### 3. Consensus Algorithm - Raft (03_consensus_raft)
- 분산 합의 알고리즘의 핵심
- Leader Election, Log Replication
- 분산 환경에서의 일관성 보장

### 4. Distributed Locking (04_distributed_locking)
- 분산 환경에서의 동기화
- Redis 기반 분산 락
- 락 타임아웃 및 재시도 전략

### 5. Message Queue (05_message_queue)
- 비동기 메시징 시스템
- Producer-Consumer 패턴
- 메시지 영속성 및 재전송

### 6. Service Discovery (06_service_discovery)
- 동적 서비스 등록 및 발견
- Health Check 및 자동 제거
- 서비스 메타데이터 관리

### 7. Circuit Breaker (07_circuit_breaker)
- 장애 전파 방지
- Closed, Open, Half-Open 상태 관리
- 자동 복구 메커니즘

### 8. Distributed Caching (08_distributed_caching)
- 분산 캐시 구현
- 캐시 일관성 유지
- Eviction 정책 (LRU, TTL)

### 9. Leader Election (09_leader_election)
- 분산 환경에서의 리더 선출
- 고가용성 보장
- Split-brain 문제 해결

### 10. Rate Limiting (10_rate_limiting)
- API 요청 제한
- Token Bucket, Leaky Bucket 알고리즘
- 분산 환경에서의 Rate Limiting

## 🚀 실습 방법

각 디렉토리는 독립적인 실습 프로젝트로 구성되어 있습니다.

```bash
# 특정 개념 실습
cd 01_client_server
go run .

# 또는 각 디렉토리의 README.md 참고
```

## 📖 학습 순서 추천

1. **01_client_server** - 기본 통신 이해
2. **02_load_balancing** - 트래픽 분산
3. **07_circuit_breaker** - 장애 대응
4. **05_message_queue** - 비동기 통신
5. **04_distributed_locking** - 동기화
6. **08_distributed_caching** - 성능 최적화
7. **10_rate_limiting** - 과부하 방지
8. **06_service_discovery** - 서비스 관리
9. **09_leader_election** - 고가용성
10. **03_consensus_raft** - 합의 알고리즘 (가장 복잡)

## 🔧 필요한 도구

- Go 1.21+
- Redis (일부 실습에 필요)
- Docker (선택사항, 멀티 인스턴스 테스트용)

## 📝 참고 자료

- [Designing Data-Intensive Applications](https://dataintensive.net/)
- [Distributed Systems for Fun and Profit](http://book.mixu.net/distsys/)
- [The Raft Consensus Algorithm](https://raft.github.io/)
