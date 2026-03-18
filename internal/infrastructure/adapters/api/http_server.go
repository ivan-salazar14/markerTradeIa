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
	hedgeController      *controllers.HedgeController
	authService          *auth.AuthService
}

func NewRouter(authSvc *auth.AuthService, monController *controllers.MonitoringController, hedgeController *controllers.HedgeController) *Router {
	return &Router{
		authController:       controllers.NewAuthController(authSvc),
		monitoringController: monController,
		hedgeController:      hedgeController,
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

		// Hedge Strategy Routes
		mux.Route("/api/hedge", func(mux chi.Router) {
			mux.Get("/strategy", r.hedgeController.GetStrategy)
			mux.Get("/stats", r.hedgeController.GetStats)
			mux.Get("/wallets", r.hedgeController.GetWallets)
			mux.Post("/wallets/connect", r.hedgeController.ConnectWallet)
			mux.Post("/wallets/disconnect", r.hedgeController.DisconnectWallet)
			mux.Post("/sync", r.hedgeController.SyncNow)
			mux.Get("/delta", r.hedgeController.GetDelta)
			mux.Get("/permissions", r.hedgeController.GetPermissions)
			mux.Get("/safe-mode", r.hedgeController.GetSafeMode)
			mux.Get("/sync-flow", r.hedgeController.GetSyncFlow)
		})
	})

	return mux
}

func StartServer(port int, handler http.Handler) error {
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("HTTP Server starting on %s\n", addr)
	return http.ListenAndServe(addr, handler)
}
