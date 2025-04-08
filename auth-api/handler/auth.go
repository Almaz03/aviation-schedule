package handler

import (
	"aviation-schedule/auth-api/middleware"
	"aviation-schedule/auth-api/model"
	"database/sql"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

var DB *sql.DB

func SetDB(db *sql.DB) {
	DB = db
}

func Register(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	err := DB.QueryRow("INSERT INTO users (username, password, role) VALUES ($1, $2, $3) RETURNING id", user.Username, string(hashed), "user").Scan(&user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "username already exists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "registered"})
}

func Login(c *gin.Context) {
	var input model.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user model.User
	err := DB.QueryRow("SELECT id, username, password, role FROM users WHERE username = $1", input.Username).Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, _ := middleware.GenerateToken(user.Username, user.Role)
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func Me(c *gin.Context) {
	token := c.GetHeader("Authorization")[7:] // Bearer ...
	username, role, err := middleware.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"username": username, "role": role})
}
