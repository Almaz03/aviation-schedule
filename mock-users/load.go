package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	authToken = "" // полученный токен
	mu        sync.RWMutex
)

const (
	loginURL = "http://localhost:8082/login"
	apiKey   = "TNUr9MZK3Sgmy5hswnJGCExvH7VacRbpFDP6YA4Wuf8dj"
	adminKey = "khyWYbSHGjxUd98J2BwR4fNPrpgXv6ztZVmDAELqCs7Kc"
)

var (
	concurrency = 100 // горутин в одной волне
	duration    = 300 // общее время работы
	interval    = 5   // интервал между волнами (сек)
	waves       = 5   // количество волн
)

type Endpoint struct {
	Name    string
	Method  string
	URL     string
	Headers map[string]string
	Body    func() []byte
}

var endpoints = []Endpoint{
	{
		Name:   "/me",
		Method: "GET",
		URL:    "http://localhost:8082/me",
		Headers: map[string]string{
			"Authorization": "$TOKEN",
		},
	},
	{
		Name:   "/flights",
		Method: "GET",
		URL:    "http://localhost:8000/flights",
		Headers: map[string]string{
			"API-Key": apiKey,
		},
	},
	{
		Name:   "/flights",
		Method: "POST",
		URL:    "http://localhost:8000/flights",
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "$TOKEN",
			"API-Key":       apiKey,
		},
		Body: randomFlightBody,
	},
	{
		Name:   "/flights/all",
		Method: "DELETE",
		URL:    "http://localhost:8000/flights/all",
		Headers: map[string]string{
			"Authorization": "$TOKEN",
			"API-Key":       apiKey,
		},
	},
	{
		Name:   "/flights/past",
		Method: "DELETE",
		URL:    "http://localhost:8000/flights/past",
		Headers: map[string]string{
			"Authorization": "$TOKEN",
			"API-Key":       apiKey,
		},
	},
	{
		Name:   "/register",
		Method: "POST",
		URL:    "http://localhost:8082/register",
		Headers: map[string]string{
			"Content-Type": "application/json",
			"API-Key":      adminKey,
		},
		Body: randomUserBody,
	},
}

func startLoad(ctx context.Context) {
	// читаем переменные окружения
	if c := os.Getenv("CONCURRENCY"); c != "" {
		if val, err := strconv.Atoi(c); err == nil {
			concurrency = val
		}
	}
	if d := os.Getenv("DURATION"); d != "" {
		if val, err := strconv.Atoi(d); err == nil {
			duration = val
		}
	}
	if w := os.Getenv("WAVES"); w != "" {
		if val, err := strconv.Atoi(w); err == nil {
			waves = val
		}
	}

	log.Println("🔐 Логинимся...")

	if err := loginAndStoreToken(); err != nil {
		log.Fatalf("❌ Не удалось залогиниться: %v", err)
	}

	log.Printf("🌀 Запуск %d волн с интервалом %d сек и %d горутинами на endpoint\n", waves, interval, concurrency)

	for wave := 1; wave <= waves; wave++ {
		select {
		case <-ctx.Done():
			log.Println("Контекст завершен, прекращаем нагрузку")
			return
		default:
			log.Printf("🌊 Волна #%d", wave)
			waveNumber.Set(float64(wave))

			var wg sync.WaitGroup
			for _, ep := range endpoints {
				for i := 0; i < concurrency; i++ {
					wg.Add(1)
					go func(e Endpoint) {
						defer wg.Done()

						reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
						defer cancel()

						var body []byte
						if e.Body != nil {
							body = e.Body()
						}

						req, err := http.NewRequestWithContext(reqCtx, e.Method, e.URL, bytes.NewBuffer(body))
						if err != nil {
							log.Printf("Ошибка запроса к %s: %v", e.Name, err)
							return
						}

						// подстановка токена
						for k, v := range e.Headers {
							if v == "$TOKEN" {
								mu.RLock()
								req.Header.Set(k, authToken)
								mu.RUnlock()
							} else {
								req.Header.Set(k, v)
							}
						}

						start := time.Now()
						resp, err := http.DefaultClient.Do(req)
						if err != nil {
							if ctx.Err() == nil {
								log.Printf("❌ %s %s [ошибка: %v]", e.Method, e.URL, err)
							}
							return
						}
						defer resp.Body.Close()

						// Добавление инкремента для ошибок 400+
						if resp.StatusCode >= 400 {
							requestErrors.WithLabelValues(e.Name).Inc()
						}

						duration := time.Since(start).Seconds()
						requestDuration.WithLabelValues(e.Name).Observe(duration)
						requestTotal.WithLabelValues(e.Name).Inc()

						if ctx.Err() == nil {
							log.Printf("→ %s %s [%d]", e.Method, e.URL, resp.StatusCode)
						}
					}(ep)
				}
			}
			wg.Wait()

			select {
			case <-ctx.Done():
				log.Println("Контекст завершен, прекращаем нагрузку")
				return
			case <-time.After(time.Duration(interval) * time.Second):
			}
		}
	}
}

func loginAndStoreToken() error {
	payload := []byte(`{"username":"admin","password":"admin"}`)

	req, err := http.NewRequest("POST", loginURL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("статус %d при логине", res.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return err
	}

	if result.Token == "" {
		return fmt.Errorf("токен не получен (пустой)")
	}

	mu.Lock()
	authToken = "Bearer " + result.Token
	mu.Unlock()

	// безопасный вывод части токена
	displayToken := result.Token
	if len(result.Token) > 12 {
		displayToken = result.Token[:12] + "..."
	}
	log.Printf("🔓 Получен токен: %s\n", displayToken)
	return nil
}

func randomUserBody() []byte {
	return []byte(fmt.Sprintf(
		`{"username":"user_%s","password":"pass_%s","role":"admin"}`,
		randomString(6), randomString(8),
	))
}

func randomFlightBody() []byte {
	now := time.Now().Add(1 * time.Hour)
	departure := now.Format(time.RFC3339)
	arrival := now.Add(2 * time.Hour).Format(time.RFC3339)

	return []byte(fmt.Sprintf(
		`{"number":"QA%d","origin":"UUEE","destination":"UUWW","departure_time":"%s","arrival_time":"%s","status":"scheduled"}`,
		rand.Intn(1000), departure, arrival,
	))
}

func randomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	letters := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
