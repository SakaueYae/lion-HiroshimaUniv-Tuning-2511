package server

import (
	"backend/internal/db"
	"backend/internal/handler"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/service"
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/riandyrn/otelchi"
)

type Server struct {
	Router *chi.Mux
}

func NewServer() (*Server, *sqlx.DB, error) {
	dbConn, err := db.InitDBConnection()
	if err != nil {
		return nil, nil, err
	}

	store := repository.NewStore(dbConn)

	authService := service.NewAuthService(store)
	orderService := service.NewOrderService(store)
	productService := service.NewProductService(store)
	robotService := service.NewRobotService(store)

	authHandler := handler.NewAuthHandler(authService)
	productHandler := handler.NewProductHandler(productService)
	orderHandler := handler.NewOrderHandler(orderService)
	robotHandler := handler.NewRobotHandler(robotService)

	userAuthMW := middleware.UserAuthMiddleware(store.SessionRepo)

	robotAPIKey := os.Getenv("ROBOT_API_KEY")
	if robotAPIKey == "" {
		log.Println("Warning: ROBOT_API_KEY is not set. Using default key 'test-robot-key'")
		robotAPIKey = "test-robot-key"
	}
	robotAuthMW := middleware.RobotAuthMiddleware(robotAPIKey)

	r := chi.NewRouter()
	r.Use(otelchi.Middleware(
		"backend-api",
		otelchi.WithChiRoutes(r),
		otelchi.WithFilter(func(req *http.Request) bool {
			return req.URL.Path != "/api/health"
		}),
	))

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	s := &Server{
		Router: r,
	}

	s.setupRoutes(authHandler, productHandler, orderHandler, robotHandler, userAuthMW, robotAuthMW)

	// セッションクリーンアップ処理を開始
	s.StartSessionCleanup(dbConn)

	return s, dbConn, nil
}

func (s *Server) setupRoutes(
	authHandler *handler.AuthHandler,
	productHandler *handler.ProductHandler,
	orderHandler *handler.OrderHandler,
	robotHandler *handler.RobotHandler,
	userAuthMW func(http.Handler) http.Handler,
	robotAuthMW func(http.Handler) http.Handler,
) {
	s.Router.Post("/api/login", authHandler.Login)

	s.Router.Route("/api/v1", func(r chi.Router) {
		r.Use(userAuthMW)
		r.Post("/product", productHandler.List)
		r.Post("/product/post", productHandler.CreateOrders)
		r.Post("/orders", orderHandler.List)
		r.Get("/image", productHandler.GetImage)
	})

	s.Router.Route("/api/robot", func(r chi.Router) {
		r.Use(robotAuthMW)
		r.Get("/delivery-plan", robotHandler.GetDeliveryPlan)
		r.Patch("/orders/status", robotHandler.UpdateOrderStatus)
	})
}

func (s *Server) Run() {
	appPort := os.Getenv("PORT")
	if appPort == "" {
		appPort = "8080"
	}

	// HTTPサーバーの詳細設定
	srv := &http.Server{
		Addr:    ":" + appPort,
		Handler: s.Router,

		// タイムアウト設定
		ReadTimeout:       10 * time.Second,  // リクエスト読み取り
		WriteTimeout:      30 * time.Second,  // レスポンス書き込み（画像配信を考慮）
		IdleTimeout:       120 * time.Second, // キープアライブ
		ReadHeaderTimeout: 5 * time.Second,   // ヘッダ読み取り

		// その他の設定
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	log.Printf("Starting server on :%s", appPort)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// StartSessionCleanup は期限切れセッションを定期的に削除
func (s *Server) StartSessionCleanup(db *sqlx.DB) {
	go func() {
		ticker := time.NewTicker(1 * time.Hour) // 1時間ごとに実行
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				result, err := db.ExecContext(ctx,
					"DELETE FROM user_sessions WHERE expires_at < NOW()")
				cancel()

				if err != nil {
					log.Printf("Session cleanup failed: %v", err)
				} else {
					rowsAffected, _ := result.RowsAffected()
					if rowsAffected > 0 {
						log.Printf("Cleaned up %d expired sessions", rowsAffected)
					}
				}
			}
		}
	}()
}
