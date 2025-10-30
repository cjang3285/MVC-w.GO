package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
    
    "github.com/gin-gonic/gin"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

var db *gorm.DB

func main() {
    // DB 연결 테스트
    if err := connectDB(); err != nil {
        log.Fatal("DB 연결 실패:", err)
    }
    log.Println("✅ PostgreSQL 연결 성공!")
    
    // Gin 라우터
    r := gin.Default()
    
    // 헬스체크
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "ok",
            "db":     "connected",
            "nginx":  "proxied",
        })
    })
    
    // Nginx 연결 테스트용
    r.GET("/api/test", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "message": "Go 서버 작동 중!",
        })
    })
    
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    log.Printf("🚀 서버 시작: :%s\n", port)
    r.Run(":" + port)
}

func connectDB() error {
    var err error
    dsn := fmt.Sprintf(
        "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
        os.Getenv("DB_HOST"),
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_NAME"),
        os.Getenv("DB_PORT"),
    )
    
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return err
    }
    
    // 연결 테스트
    sqlDB, err := db.DB()
    if err != nil {
        return err
    }
    
    return sqlDB.Ping()
}
