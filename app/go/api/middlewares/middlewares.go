package middlewares

import (
	"errors"
	"go-ddd-cqrs-example/api/auth"
	"go-ddd-cqrs-example/api/responses"
	"go-ddd-cqrs-example/api/server"
	"log"
	"net/http"
)

// SetMiddlewareJSON sets server response type to json.
func SetMiddlewareJSON(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers","Content-Type,access-control-allow-origin, access-control-allow-headers")
		next(w, r)
		}
}

// SetMiddlewareAuthentication sets auth for the server.
func SetMiddlewareAuthentication(server server.Server, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		valid, err := auth.CheckJWTTokenValidity(server, r)
		if valid == false {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
			return
		}

		if err != nil {
			log.Panic(err)
		}

		next(w, r)
	}
}
