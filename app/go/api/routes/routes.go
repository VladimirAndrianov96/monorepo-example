package routes

import (
	"go-ddd-cqrs-example/api/controllers/login_controller"
	user_controller "go-ddd-cqrs-example/api/controllers/user"
	"go-ddd-cqrs-example/api/middlewares"
	"go-ddd-cqrs-example/api/server"
)

func InitializeRoutes(s *server.Server) {
	// Auth routes
	s.Router.HandleFunc("/api/login", middlewares.SetMiddlewareJSON(login_controller.Login(s))).Methods("POST")
	s.Router.HandleFunc("/api/register", middlewares.SetMiddlewareJSON(user_controller.Register(s))).Methods("POST")

	//// User routes
	s.Router.HandleFunc("/api/deactivate/current", middlewares.SetMiddlewareJSON(middlewares.SetMiddlewareAuthentication(*s, user_controller.Deactivate(s)))).Methods("POST")
	s.Router.HandleFunc("/api/activate/current", middlewares.SetMiddlewareJSON(middlewares.SetMiddlewareAuthentication(*s, user_controller.Activate(s)))).Methods("POST")
}
