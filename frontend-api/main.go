package main

import (
	"log"
	"net/http"
)

func main() {
	// Обслуживаем все файлы из папки "static"
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Перенаправляем на login.html, если пользователь заходит на главную страницу
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/login.html", http.StatusFound)
	})

	// Логируем информацию о запуске
	log.Println("Frontend доступен на http://localhost:8083")

	// Запуск сервера на порту 8083
	log.Fatal(http.ListenAndServe(":8083", nil))
}
