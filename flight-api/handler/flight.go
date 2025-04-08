package handler

import (
	"aviation-schedule/flight-api/model"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"time"
)

var DB *sql.DB

func SetDB(database *sql.DB) {
	DB = database
}

func GetFlights(c *gin.Context) {
	rows, err := DB.Query("SELECT id, number, origin, destination, departure_time, arrival_time, status FROM flights")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var flights []model.Flight
	for rows.Next() {
		var f model.Flight
		_ = rows.Scan(&f.ID, &f.Number, &f.Origin, &f.Destination, &f.DepartureTime, &f.ArrivalTime, &f.Status)
		flights = append(flights, f)
	}

	c.JSON(http.StatusOK, flights)
}

func CreateFlight(c *gin.Context) {
	var f model.Flight
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := DB.QueryRow("INSERT INTO flights (number, origin, destination, departure_time, arrival_time, status) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		f.Number, f.Origin, f.Destination, f.DepartureTime, f.ArrivalTime, f.Status).Scan(&f.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, f)
}

func UpdateFlight(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var f model.Flight
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := DB.Exec("UPDATE flights SET number=$1, origin=$2, destination=$3, departure_time=$4, arrival_time=$5, status=$6 WHERE id=$7",
		f.Number, f.Origin, f.Destination, f.DepartureTime, f.ArrivalTime, f.Status, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func DeleteFlight(c *gin.Context) {
	id := c.Param("id")
	_, err := DB.Exec("DELETE FROM flights WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
func DeletePastFlights(c *gin.Context) {
	// Получаем текущее время в UTC
	now := time.Now().UTC()

	// Выполняем запрос, преобразуя строку в timestamp на стороне БД
	result, err := DB.Exec(`
        DELETE FROM flights 
        WHERE (arrival_time::timestamp with time zone) < $1::timestamp with time zone
    `, now.Format(time.RFC3339))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка при удалении прошедших рейсов: " + err.Error(),
		})
		return
	}

	// Получаем количество удаленных записей
	rowsAffected, _ := result.RowsAffected()

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Удалено %d прошедших рейсов", rowsAffected),
	})
}
func DeleteAllFlights(c *gin.Context) {
	// Пытаемся выполнить TRUNCATE, который удаляет все строки из таблицы
	_, err := DB.Exec("TRUNCATE TABLE flights RESTART IDENTITY CASCADE")
	if err != nil {
		log.Println("Ошибка при удалении всех рейсов:", err) // Логируем ошибку для диагностики
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении всех рейсов"})
		return
	}

	// Получаем количество удаленных строк
	// `TRUNCATE` не возвращает `RowsAffected`, поэтому этот шаг не потребуется.
	log.Println("Все рейсы успешно удалены")

	// Возвращаем успешный ответ
	c.JSON(http.StatusOK, gin.H{"message": "Все рейсы успешно удалены"})
}
