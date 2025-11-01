#!/bin/bash

echo "======================================"
echo "라즈베리파이 메모리 모니터링"
echo "======================================"
echo ""

# 1. 간단한 메모리 확인
echo "📊 1. 기본 메모리 정보:"
free -h
echo ""

# 2. 더 보기 좋게
echo "📊 2. 메모리 사용률 (%):"
free | grep Mem | awk '{printf "사용중: %.1f%%\n", $3/$2 * 100.0}'
echo ""

# 3. 실시간 모니터링 (top 한번만)
echo "📊 3. 프로세스별 메모리 사용 (상위 10개):"
ps aux --sort=-%mem | head -11
echo ""

# 4. Redis 메모리 확인 (Redis가 실행중이면)
echo "📊 4. Redis 메모리 사용량:"
if command -v redis-cli &> /dev/null; then
    redis-cli info memory | grep used_memory_human
else
    echo "Redis가 설치되지 않음"
fi
echo ""

# 5. PostgreSQL 메모리 확인
echo "📊 5. PostgreSQL 프로세스:"
ps aux | grep postgres | grep -v grep
echo ""

# 6. 시스템 전체 정보
echo "📊 6. 시스템 정보:"
cat /proc/meminfo | grep -E 'MemTotal|MemFree|MemAvailable|Cached|SwapTotal|SwapFree'
