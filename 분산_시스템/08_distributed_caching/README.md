# Distributed Caching (분산 캐싱)

## 개념

분산 캐시는 여러 서버에서 공유하는 캐시 시스템으로,
데이터베이스 부하를 줄이고 응답 속도를 향상시킵니다.

## 주요 특징

1. **빠른 접근**: 메모리 기반 저장으로 밀리초 단위 응답
2. **확장성**: 여러 캐시 노드로 데이터 분산
3. **일관성**: 캐시와 DB 간 데이터 동기화
4. **TTL**: 자동 만료로 메모리 관리

## Eviction 정책

- **LRU (Least Recently Used)**: 가장 오래 사용되지 않은 항목 제거
- **LFU (Least Frequently Used)**: 가장 적게 사용된 항목 제거
- **TTL (Time To Live)**: 설정된 시간 후 자동 제거

## 구현 내용

- LRU 기반 캐시
- TTL 자동 만료
- Thread-safe 구현
- Cache Hit/Miss 통계
- Cache-Aside 패턴

## 실행 방법

```bash
go run main.go
```

## 학습 포인트

- 캐시의 필요성
- LRU 알고리즘 이해
- TTL 관리
- 캐시 일관성 문제
- 실제 도구: Redis, Memcached
