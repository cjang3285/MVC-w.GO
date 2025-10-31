FROM golang:1.21-alpine

WORKDIR /app

# 의존성 먼저 복사 (캐싱)
COPY go.mod go.sum ./
RUN go mod download

# 소스 복사
COPY . .

# 빌드
RUN go build -o main .

# 실행
CMD ["./main"]
