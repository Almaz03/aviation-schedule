package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

const adminAPIKey = "khyWYbSHGjxUd98J2BwR4fNPrpgXv6ztZVmDAELqCs7Kc"
const userAPIKey = "TNUr9MZK3Sgmy5hswnJGCExvH7VacRbpFDP6YA4Wuf8dj"

func main() {
	var err error
	db, err = sql.Open("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=flights sslmode=disable")
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:8083") // Указываем ваш фронтенд
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Api-Key")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	r.POST("/login", LoginHandler)
	r.POST("/register", RegisterHandler)
	r.GET("/me", MeHandler)

	log.Println("Auth API запущен на порту 8082")
	r.Run(":8082")
}

func LoginHandler(c *gin.Context) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	var storedHashedPassword, role string
	err := db.QueryRow("SELECT password, role FROM users WHERE username = $1", creds.Username).
		Scan(&storedHashedPassword, &role)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(creds.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		return
	}

	token := fmt.Sprintf("%s-token", creds.Username)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"role":  role,
	})
}

func RegisterHandler(c *gin.Context) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	apiKey := c.GetHeader("Api-Key")

	// Проверяем API ключ
	if apiKey != adminAPIKey && apiKey != userAPIKey {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid API key"})
		return
	}

	if err := c.BindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Если ключ администратора, то можем создать и админа
	if apiKey == adminAPIKey && creds.Role != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "admin key can only create admin users"})
		return
	}

	// Если ключ обычного пользователя, то роль только user
	if apiKey == userAPIKey && creds.Role != "user" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user key can only create user accounts"})
		return
	}

	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Добавляем пользователя в БД
	_, err = db.Exec("INSERT INTO users (username, password, role) VALUES ($1, $2, $3)",
		creds.Username, string(hashedPassword), creds.Role)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot register user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "register ok"})
}

func MeHandler(c *gin.Context) {
	token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	username := strings.TrimSuffix(token, "-token")

	var user struct {
		Username string
		Role     string
	}

	err := db.QueryRow("SELECT username, role FROM users WHERE username = $1", username).
		Scan(&user.Username, &user.Role)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"username": user.Username,
		"role":     user.Role,
	})
}
