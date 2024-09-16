package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	_ "github.com/go-sql-driver/mysql"
)

var db  *gorm.DB
var rdb *redis.Client
var jwtSecret []byte

// 初始化資料庫連線和 Redis 連線
func init() {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	redisAddr := os.Getenv("REDIS_ADDR")
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))

	var err error
	// 构建 DSN（数据源名称）
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbPort, dbName)

    // 使用 GORM 连接 MySQL 数据库
    db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("無法連接資料庫: %v", err)
    }

    // 测试数据库连接
    sqlDB, err := db.DB()
    if err != nil {
        log.Fatalf("無法獲取原始數據庫連接: %v", err)
    }

    err = sqlDB.Ping()
    if err != nil {
        log.Fatalf("無法連接資料庫: %v", err)
    }

	rdb = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("無法連接 Redis: %v", err)
	}

	log.Println("成功連接資料庫和 Redis")
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/v1/item", ListItemHandler).Methods("GET")
	r.HandleFunc("/v1/item", CreateItemHandler).Methods("POST")
	r.HandleFunc("/v1/login", loginHandler).Methods("POST")
	r.HandleFunc("/v1/register", registerHandler).Methods("POST")

	log.Println("伺服器啟動於 8000 端口...")
	log.Fatal(http.ListenAndServe(":8000", r))
}
type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
}
// 註冊 API
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只允許 POST 請求", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// 加密密碼
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "密碼加密失敗", http.StatusInternalServerError)
		return
	}

	// 使用 GORM 插入新使用者
	user := User{Username: username, Password: string(hashedPassword)}
	result := db.Create(&user)
	if result.Error != nil {
		http.Error(w, "無法儲存使用者資料", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(`{"status":"success","description":"user registered"}`))
}


// 登入 API 並生成 JWT
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只允許 POST 請求", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// 使用 GORM 查詢使用者
	var user User
	result := db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		http.Error(w, "使用者不存在", http.StatusUnauthorized)
		return
	}

	// 比對加密後的密碼
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		w.Write([]byte(`{"status":"failed","description":"wrong password"}`))
		return
	}

	// 生成 JWT token
	token, err := generateJWT(username)
	if err != nil {
		http.Error(w, "無法生成 JWT", http.StatusInternalServerError)
		return
	}

	// 儲存 JWT 到 Redis
	err = rdb.Set(context.Background(), username, token, 1*time.Hour).Err()
	if err != nil {
		http.Error(w, "無法儲存 JWT 到 Redis", http.StatusInternalServerError)
		return
	}

	// 回傳 JWT 給使用者
	response := map[string]string{
		"status":       "success",
		"description":  "ok",
		"access_token": token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}


// 生成 JWT token
func generateJWT(username string) (string, error) {
	// 定義 token 的過期時間
	expirationTime := time.Now().Add(1 * time.Hour)

	// 建立 claims（包括使用者名稱和到期時間）
	claims := &jwt.RegisteredClaims{
		Subject:   username,
		ExpiresAt: jwt.NewNumericDate(expirationTime),
	}

	// 使用密鑰來簽署 token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}