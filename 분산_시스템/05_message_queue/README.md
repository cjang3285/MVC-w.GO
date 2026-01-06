# Message Queue (메시지 큐)

## 개념

Message Queue는 비동기 통신을 위한 핵심 컴포넌트로, 서비스 간 결합도를 낮추고
확장성을 높이는 데 중요한 역할을 합니다.

## 주요 특징

1. **비동기 처리**: Producer와 Consumer가 독립적으로 동작
2. **버퍼링**: 트래픽 급증 시 메시지를 큐에 저장
3. **확장성**: Consumer를 늘려 처리량 증가 가능
4. **신뢰성**: 메시지 영속성 및 재전송 보장

## 구현 패턴

- **Producer**: 메시지를 큐에 전송
- **Queue**: 메시지를 저장하고 관리
- **Consumer**: 큐에서 메시지를 가져와 처리

## 구현 내용

- 인메모리 메시지 큐
- Producer-Consumer 패턴
- 메시지 ACK 메커니즘
- 여러 Consumer의 동시 처리

## 실행 방법

```bash
go run main.go
```

## 학습 포인트

- 비동기 통신의 장점
- Producer-Consumer 패턴
- 메시지 영속성 (실제 환경: RabbitMQ, Kafka)
- At-least-once vs At-most-once 전달 보장
