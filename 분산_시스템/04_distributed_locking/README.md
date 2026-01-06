# Distributed Locking (분산 락)

## 개념

분산 환경에서 여러 프로세스가 공유 자원에 접근할 때 동기화를 보장하는 메커니즘입니다.
분산 락이 없으면 Race Condition이 발생하여 데이터 불일치가 발생할 수 있습니다.

## 주요 특징

1. **Mutual Exclusion**: 한 번에 하나의 프로세스만 락 획득
2. **Deadlock Free**: 락을 획득한 프로세스가 죽어도 자동 해제
3. **Fault Tolerance**: 분산 환경에서도 안정적으로 동작

## 구현 방식

이 구현에서는 인메모리 방식으로 간단한 분산 락을 구현했습니다.
실제 프로덕션 환경에서는 Redis나 etcd 같은 도구를 사용합니다.

### 구현 내용
- Lock/Unlock 메커니즘
- TTL (Time To Live) 기반 자동 해제
- 락 획득 재시도 로직
- 동시성 테스트

## 실행 방법

```bash
go run main.go
```

## 학습 포인트

- 분산 환경에서의 동기화
- Race Condition 방지
- TTL의 중요성
- Redis를 이용한 실제 구현 방법
