package server

import (
	"net/http"
)

// AccessControlMiddleware добавляет заголовок "Access-Control-Allow-Origin" в ответ.
func AccessControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Добавляем заголовок для разрешения кросс-доменных запросов
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Если это запрос OPTIONS, сразу отправляем пустой ответ
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Передаем управление следующему обработчику
		next.ServeHTTP(w, r)
	})
}
