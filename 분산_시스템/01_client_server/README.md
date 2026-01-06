# Client-Server 모델

## 개념

Client-Server 모델은 분산 시스템의 가장 기본적인 아키텍처입니다.
- **Server**: 서비스를 제공하는 프로세스
- **Client**: 서비스를 요청하는 프로세스

## 구현 내용

1. HTTP 기반 서버 구현
2. 클라이언트에서 서버로 요청 전송
3. 서버에서 응답 반환
4. 동시 요청 처리

## 실행 방법

```bash
# 서버 실행
go run server.go

# 다른 터미널에서 클라이언트 실행
go run client.go
```

## 학습 포인트

- HTTP 프로토콜 이해
- 동시성 처리 (Goroutine)
- 요청-응답 패턴
- JSON 직렬화/역직렬화
