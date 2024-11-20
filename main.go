package main

import (
	"fmt"
	"net/http"
)

func enableCORS(w http.ResponseWriter, r *http.Request) {
	// Разрешаем доступ с любого источника
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// Разрешаем определенные методы
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	// Разрешаем заголовки
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Если это preflight запрос (OPTIONS), сразу отправляем 200 OK
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Ваш код обработки запроса на /login
	fmt.Fprintf(w, "Login successful!")
}

func main() {
	// Применяем CORS для всех маршрутов
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		loginHandler(w, r)
	})

	// Применяем CORS для других маршрутов (если необходимо)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		w.Write([]byte("Hello, World!"))
	})

	// Запуск сервера на порту 8080
	http.ListenAndServe(":8080", nil)
}
