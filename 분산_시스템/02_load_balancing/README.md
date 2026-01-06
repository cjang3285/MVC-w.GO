# Load Balancing (부하 분산)

## 개념

Load Balancer는 들어오는 네트워크 트래픽을 여러 서버에 분산시켜
시스템의 가용성과 성능을 향상시키는 핵심 컴포넌트입니다.

## 구현 알고리즘

1. **Round Robin**: 순차적으로 서버 선택
2. **Least Connections**: 연결 수가 가장 적은 서버 선택
3. **Health Check**: 장애 서버 자동 제외

## 구현 내용

- 로드 밸런서 (여러 백엔드 서버로 요청 분산)
- 백엔드 서버 (포트 8081, 8082, 8083)
- Health Check 메커니즘
- 동적 서버 추가/제거

## 실행 방법

```bash
# 로드 밸런서 및 백엔드 서버 실행
go run main.go

# 다른 터미널에서 클라이언트 실행
go run client.go
```

## 학습 포인트

- 부하 분산 알고리즘 비교
- 서버 헬스 체크
- 고가용성 설계
- Reverse Proxy 개념
