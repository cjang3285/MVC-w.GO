.PHONY: up down logs build clean test

# 전체 시작
up:
	docker-compose up -d
	@echo "✅ 서버 시작 완료!"
	@echo "🌐 Nginx: http://localhost"
	@echo "🔌 API: http://localhost:8080"
	@echo "💾 DB: localhost:5432"

# 중지
down:
	docker-compose down

# 로그 보기
logs:
	docker-compose logs -f

# 빌드
build:
	docker-compose build

# 완전 삭제 (DB 데이터 포함)
clean:
	docker-compose down -v

# 연결 테스트
test:
	@echo "🧪 연결 테스트..."
	@echo "\n1️⃣ Nginx 테스트:"
	@curl -s http://localhost || echo "❌ Nginx 연결 실패"
	@echo "\n2️⃣ API 직접 테스트:"
	@curl -s http://localhost:8080/api/test || echo "❌ API 연결 실패"
	@echo "\n3️⃣ Nginx → API 테스트:"
	@curl -s http://localhost/api/test || echo "❌ Nginx → API 연결 실패"
	@echo "\n4️⃣ 헬스체크:"
	@curl -s http://localhost/health || echo "❌ 헬스체크 실패"
	@echo "\n✅ 테스트 완료!"

# 재시작
restart: down up
