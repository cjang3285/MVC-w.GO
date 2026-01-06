# Circuit Breaker (서킷 브레이커)

## 개념

Circuit Breaker는 장애가 발생한 서비스로의 요청을 차단하여
전체 시스템의 연쇄 장애를 방지하는 디자인 패턴입니다.

전기 회로의 차단기(Circuit Breaker)에서 이름을 따왔습니다.

## 세 가지 상태

1. **Closed (정상)**: 모든 요청이 정상적으로 처리됨
2. **Open (차단)**: 실패율이 임계값을 넘어 요청을 즉시 차단
3. **Half-Open (반열림)**: 일부 요청을 허용하여 복구 확인

## 상태 전환

```
Closed --[실패율 > 임계값]--> Open
Open --[타임아웃 후]--> Half-Open
Half-Open --[성공]--> Closed
Half-Open --[실패]--> Open
```

## 장점

- 장애 전파 방지
- 빠른 실패 (Fail Fast)
- 시스템 복구 시간 확보
- 자동 복구 메커니즘

## 구현 내용

- 3가지 상태 관리
- 실패율 기반 상태 전환
- 자동 복구 시도
- 타임아웃 설정

## 실행 방법

```bash
go run main.go
```

## 학습 포인트

- Fail Fast 패턴
- 장애 격리 (Fault Isolation)
- 자동 복구 메커니즘
- 실제 라이브러리: Hystrix, resilience4j
