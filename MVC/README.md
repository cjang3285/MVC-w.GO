# MVC project
Go + Nginx + PostgreSQL 연결 테스트

## 빠른 시작

### 요구사항
- Docker
- Docker Compose

### 설치 & 실행
```bash
# 1. 클론
git clone https://github.com/chanwookjang/MVC.git
cd MVC

# 2. 실행 (모든 의존성 자동 설치)
make up

# 또는
docker-compose up -d
```

### 연결 테스트
```bash
# 자동 테스트
make test

# 수동 테스트
curl http://localhost              # Nginx
curl http://localhost/api/test     # Nginx → Go
curl http://localhost/health       # 전체 상태
```

## 명령어
```bash
make up      # 시작
make down    # 중지
make logs    # 로그 보기
make test    # 연결 테스트
make clean   # 완전 삭제
make restart # 재시작
```

## 구조
```
Nginx (80) → Go API (8080) → PostgreSQL (5432)
```

## 환경변수

`.env` 파일 수정:
```env
DB_USER=streamuser
DB_PASSWORD=yourpassword
DB_NAME=streamdb
```
```

---

## 9. .gitignore
```
# Binaries
main
*.exe

# Go
go.sum

# Docker
.env.local

# IDE
.vscode/
.idea/

# OS
.DS_Store
