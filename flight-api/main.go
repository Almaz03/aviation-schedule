package main

import (
	"aviation-schedule/flight-api/db"
	"aviation-schedule/flight-api/handler"
	"github.com/gin-gonic/gin"
	"net/http"
)

const API_KEY = "TNUr9MZK3Sgmy5hswnJGCExvH7VacRbpFDP6YA4Wuf8dj"

func main() {
	database := db.Init()
	handler.SetDB(database)

	r := gin.Default()

	// ✅ CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Api-Key")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// Получаем API ключ из заголовка
		apiKey := c.GetHeader("API-Key")

		// Проверяем, соответствует ли API ключ
		if apiKey != API_KEY {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Invalid API Key"})
			c.Abort()
			return
		}

		// Проверка на OPTIONS запросы
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Роуты для работы с рейсами
	r.GET("/flights", handler.GetFlights)
	r.POST("/flights", handler.CreateFlight)
	r.PUT("/flights/:id", handler.UpdateFlight)
	r.DELETE("/flights/:id", handler.DeleteFlight)
	r.DELETE("/flights/past", handler.DeletePastFlights)
	r.DELETE("/flights/all", handler.DeleteAllFlights)

	// Запуск сервера
	r.Run(":8000")
}
