package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ivan-salazar14/markerTradeIa/internal/application/services/auth"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/api/controllers"
	apiMiddleware "github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/api/middleware"
)

type Router struct {
	authController       *controllers.AuthController
	monitoringController *controllers.MonitoringController
	authService          *auth.AuthService
}

func NewRouter(authSvc *auth.AuthService, monController *controllers.MonitoringController) *Router {
	return &Router{
		authController:       controllers.NewAuthController(authSvc),
		monitoringController: monController,
		authService:          authSvc,
	}
}

func (r *Router) Init() http.Handler {
	mux := chi.NewRouter()

	// Generic Middleware
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)

	// Public Routes
	mux.Post("/auth/login", r.authController.Login)
	mux.Post("/auth/refresh", r.authController.Refresh)

	// Protected Routes (M2M or User)
	mux.Group(func(mux chi.Router) {
		mux.Use(apiMiddleware.AuthMiddleware(r.authService))

		mux.Get("/api/v1/pools", r.monitoringController.GetPools)
	})

	return mux
}

func StartServer(port int, handler http.Handler) error {
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("HTTP Server starting on %s\n", addr)
	return http.ListenAndServe(addr, handler)
}
