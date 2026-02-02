package middleware

import (
	"bytes"
	"io"
	"log"
	"net/http"
)

func DebugMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Логируем детали запроса
		log.Printf("DEBUG: Method=%s, URL=%s, RemoteAddr=%s",
			r.Method, r.URL.String(), r.RemoteAddr)
		log.Printf("DEBUG: Headers: %v", r.Header)

		// Читаем и логируем body (только для POST/PUT)
		if r.Method == "POST" || r.Method == "PUT" {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Printf("DEBUG: Error reading body: %v", err)
			} else {
				log.Printf("DEBUG: Body: %s", string(body))
				// Восстанавливаем body для следующих handlers
				r.Body = io.NopCloser(bytes.NewBuffer(body))
			}
		}

		next.ServeHTTP(w, r)
	})
}
