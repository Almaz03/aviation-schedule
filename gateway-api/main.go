package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const API_KEY = "TNUr9MZK3Sgmy5hswnJGCExvH7VacRbpFDP6YA4Wuf8dj"
const AVIATIONSTACK_API = "fb44e5a1b5ad5e0838894ba916c1c241"

type Flight struct {
	FlightDate   string `json:"flight_date"`
	FlightNumber struct {
		IATA string `json:"iata"` // flight_iata
	} `json:"flight"`
	Departure struct {
		Airport   string `json:"airport"`
		Scheduled string `json:"scheduled"`
	} `json:"departure"`
	Arrival struct {
		Airport   string `json:"airport"`
		Scheduled string `json:"scheduled"`
	} `json:"arrival"`
	Status string `json:"flight_status"`
}

func GetFlightsFromAirport(icaoCode string) ([]Flight, error) {
	url := fmt.Sprintf("https://api.aviationstack.com/v1/flights?access_key=%s&dep_icao=%s", AVIATIONSTACK_API, icaoCode)

	resp, err := http.Get(url)
	if err != nil {
		log.Println("❌ Ошибка при запросе к API:", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var flightsResponse struct {
		Data []Flight `json:"data"`
	}

	err = json.Unmarshal(body, &flightsResponse)
	if err != nil {
		log.Println("❌ Ошибка при разборе JSON:", err)
		return nil, err
	}

	return flightsResponse.Data, nil
}

func postToFlightAPI(f Flight) error {
	// Формируем тело запроса
	body := map[string]interface{}{
		"number":         f.FlightNumber.IATA,
		"origin":         f.Departure.Airport,
		"destination":    f.Arrival.Airport,
		"departure_time": f.Departure.Scheduled,
		"arrival_time":   f.Arrival.Scheduled,
		"status":         f.Status,
	}

	// Преобразуем тело в JSON
	jsonData, err := json.Marshal(body)
	if err != nil {
		log.Println("❌ Ошибка при марштализации данных:", err)
		return err
	}

	// Создаем запрос в flight-api для добавления рейса
	req, err := http.NewRequest("POST", "http://localhost:8000/flights", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("❌ Ошибка при создании запроса:", err)
		return err
	}

	// Добавляем необходимые заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("API-Key", API_KEY)                    // API ключ для аутентификации
	req.Header.Set("Authorization", "Bearer admin-token") // Замените на актуальный токен, если нужно

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("❌ Ошибка при отправке запроса:", err)
		return err
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		responseText, _ := io.ReadAll(resp.Body)
		log.Printf("⚠️ Ошибка %d: %s\n", resp.StatusCode, responseText)
		return fmt.Errorf("Ошибка при отправке в API: %s", responseText)
	}

	log.Println("✅ Рейс успешно добавлен в систему.")
	return nil
}

// Проверка API ключа
func checkAPIKey(r *http.Request) bool {
	apiKey := r.Header.Get("API-Key")
	return apiKey == API_KEY
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Разрешаем CORS
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8083") // разрешаем доступ с фронтенда на localhost:8083
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, API-Key")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

	// Проверка на OPTIONS запросы
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent) // Отвечаем на OPTIONS запрос
		return
	}

	// Проверка API ключа
	if !checkAPIKey(r) {
		http.Error(w, "Forbidden: Invalid API Key", http.StatusForbidden)
		return
	}

	icao := r.URL.Query().Get("icao")
	if icao == "" {
		http.Error(w, "Не указан параметр ?icao=XXX", http.StatusBadRequest)
		return
	}

	fmt.Printf("📡 Загрузка рейсов с ICAO: %s...\n", icao)
	flights, err := GetFlightsFromAirport(icao)
	if err != nil {
		http.Error(w, "Ошибка при получении данных с внешнего API", http.StatusInternalServerError)
		return
	}

	success := 0
	fail := 0
	for _, f := range flights {
		err := postToFlightAPI(f)
		if err == nil {
			success++
		} else {
			fail++
		}
	}

	log.Printf("✅ Успешно: %d, ❌ Ошибки: %d (из %d рейсов)\n", success, fail, len(flights))

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "✅ Загружено: %d рейсов\n❌ Ошибки: %d\n📦 Всего обработано: %d\n\nICAO: %s\n",
		success, fail, len(flights), icao)
}

func main() {
	http.HandleFunc("/parse", handler)
	fmt.Println("🚀 Gateway запущен на http://localhost:8084")
	log.Fatal(http.ListenAndServe(":8084", nil))
}
