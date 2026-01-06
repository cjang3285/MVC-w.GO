# Consensus Algorithm - Raft (간소화 버전)

## 개념

Raft는 분산 시스템에서 여러 노드가 하나의 값에 합의하도록 하는 알고리즘입니다.
이해하기 쉽게 설계되었으며, 리더 선출과 로그 복제를 통해 일관성을 보장합니다.

## Raft의 핵심 요소

1. **Leader Election**: 리더를 선출하는 과정
2. **Log Replication**: 리더가 로그를 팔로워에게 복제
3. **Safety**: 시스템의 일관성 보장

### 노드 상태
- **Follower**: 리더의 명령을 따름
- **Candidate**: 리더 선출에 참여
- **Leader**: 모든 클라이언트 요청을 처리

## 구현 내용 (간소화)

- 기본 리더 선출 메커니즘
- 하트비트를 통한 리더 유지
- Term(임기) 관리
- 투표 시스템

## 실행 방법

```bash
# 3개의 노드로 Raft 클러스터 시작
go run main.go
```

## 학습 포인트

- 분산 합의의 중요성
- Leader Election 프로세스
- Timeout 기반 상태 전환
- 과반수 투표의 의미
