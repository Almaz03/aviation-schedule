package main

import (
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/login.html", http.StatusFound)
	})

	log.Println("🚀🚀🚀🚀🚀 Frontend доступен на http://localhost:8083")

	log.Fatal(http.ListenAndServe(":8083", nil))
}
