package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	apikey_handler "llm-gateway/internal/apikey/handler"
	apikey_repo "llm-gateway/internal/apikey/repository"
	apikey_usecase "llm-gateway/internal/apikey/usecase"

	billing_handler "llm-gateway/internal/billing/handler"
	billing_repo "llm-gateway/internal/billing/repository"
	billing_usecase "llm-gateway/internal/billing/usecase"

	channel_handler "llm-gateway/internal/channel/handler"
	channel_repo "llm-gateway/internal/channel/repository"
	channel_usecase "llm-gateway/internal/channel/usecase"

	model_handler "llm-gateway/internal/model/handler"
	model_repo "llm-gateway/internal/model/repository"
	model_usecase "llm-gateway/internal/model/usecase"

	"llm-gateway/internal/relay"
	relay_handler "llm-gateway/internal/relay/handler"
	relay_usecase "llm-gateway/internal/relay/usecase"

	usage_handler "llm-gateway/internal/usage/handler"
	usage_repo "llm-gateway/internal/usage/repository"
	usage_usecase "llm-gateway/internal/usage/usecase"

	"llm-gateway/internal/shared/config"
	"llm-gateway/internal/shared/crypto"
	"llm-gateway/internal/shared/database"
	"llm-gateway/internal/shared/event"
	"llm-gateway/internal/shared/jwt"
	"llm-gateway/internal/shared/middleware"
	"llm-gateway/internal/shared/redis"

	user_handler "llm-gateway/internal/user/handler"
	user_repo "llm-gateway/internal/user/repository"
	user_usecase "llm-gateway/internal/user/usecase"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg := config.Load()

	if err := database.Init(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := redis.Init(cfg); err != nil {
		log.Fatalf("Failed to initialize redis: %v", err)
	}
	defer redis.Close()

	db := database.GetDB()
	rdb := redis.GetClient()
	keyCrypto := crypto.NewChaCha20Poly1305Crypto(cfg.ChaCha20Poly1305Key)

	token_manager := jwt.NewTokenManager(cfg.JWTSecret, cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry)

	// User module
	user_repository := user_repo.NewUserRepository(db)
	auth_usecase := user_usecase.NewAuthUsecase(user_repository, token_manager)
	user_usecase_instance := user_usecase.NewUserUsecase(user_repository)
	auth_handler := user_handler.NewAuthHandler(auth_usecase)
	user_handler := user_handler.NewUserHandler(user_usecase_instance)

	// API Key module
	apikey_repository := apikey_repo.NewApiKeyRepository(db)
	apikey_usecase_instance := apikey_usecase.NewKeyUsecase(apikey_repository)
	apikey_handler := apikey_handler.NewKeyHandler(apikey_usecase_instance)

	// Model management module
	model_repository := model_repo.NewModelRepository(db)
	model_usecase_instance := model_usecase.NewModelUsecase(model_repository)
	model_handler := model_handler.NewModelHandler(model_usecase_instance)

	// Channel module
	channel_repository := channel_repo.NewChannelRepository(db)
	channel_usecase_instance := channel_usecase.NewChannelUsecase(channel_repository, model_repository, keyCrypto)
	channel_handler := channel_handler.NewChannelHandler(channel_usecase_instance)

	// Billing module
	transaction_repository := billing_repo.NewTransactionRepository(db)
	request_log_repository := billing_repo.NewRequestLogRepository(db)
	user_repository_billing := billing_repo.NewUserRepository(db)
	apikey_repository_billing := billing_repo.NewApiKeyRepository(db)
	redeem_code_repository := billing_repo.NewRedeemCodeRepository(db)

	billing_usecase_instance := billing_usecase.NewBillingUsecase(
		user_repository_billing,
		apikey_repository_billing,
		transaction_repository,
		request_log_repository,
		model_repository,
		rdb,
	)
	redeem_usecase := billing_usecase.NewRedeemUsecase(
		redeem_code_repository,
		user_repository_billing,
		transaction_repository,
	)
	billing_handler := billing_handler.NewBillingHandler(redeem_usecase, transaction_repository)

	// Usage module
	usage_repository := usage_repo.NewRequestLogRepository(db)
	usage_usecase_instance := usage_usecase.NewUsageUsecase(usage_repository)
	usage_handler := usage_handler.NewUsageHandler(usage_usecase_instance)

	// Relay module
	channel_selector := relay.NewChannelSelector(rdb)
	event_publisher := event.NewPublisher(rdb)
	relay_usecase_instance := relay_usecase.NewRelayUsecase(
		channel_selector,
		billing_usecase_instance,
		channel_repository,
		event_publisher,
	)
	relay_handler := relay_handler.NewRelayHandler(relay_usecase_instance, model_repository, channel_repository)

	r := chi.NewRouter()

	r.Use(middleware.CORS())
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Public routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", auth_handler.Register)
			r.Post("/login", auth_handler.Login)
			r.Post("/refresh", auth_handler.RefreshToken)

			r.Group(func(r chi.Router) {
				r.Use(middleware.Auth(token_manager))
				r.Get("/me", auth_handler.GetMe)
				r.Put("/password", auth_handler.ChangePassword)
			})
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(token_manager))

			// API Key routes
			r.Route("/api-keys", func(r chi.Router) {
				r.Get("/", apikey_handler.ListKeys)
				r.Post("/", apikey_handler.CreateKey)
				r.Put("/{id}", apikey_handler.UpdateKey)
				r.Post("/{id}/toggle", apikey_handler.ToggleKey)
				r.Delete("/{id}", apikey_handler.DeleteKey)
			})

			r.Get("/models/plaza", channel_handler.ListModelPlaza)

			// Usage routes
			r.Route("/usage", func(r chi.Router) {
				r.Get("/overview", usage_handler.GetOverview)
				r.Get("/stats", usage_handler.GetStats)
				r.Get("/logs", usage_handler.GetLogs)
			})

			// Billing routes
			r.Route("/billing", func(r chi.Router) {
				r.Post("/redeem", billing_handler.Redeem)
				r.Get("/transactions", billing_handler.GetTransactions)
			})
		})

		// Admin routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(token_manager))
			r.Use(middleware.AdminOnly())

			r.Route("/admin", func(r chi.Router) {
				// User management
				r.Route("/users", func(r chi.Router) {
					r.Get("/", user_handler.ListUsers)
					r.Post("/", user_handler.CreateUser)
					r.Get("/channel-template", user_handler.ListChannelTemplate)
					r.Put("/channel-template", user_handler.SaveChannelTemplate)
					r.Post("/channel-template/apply-all", user_handler.ApplyChannelTemplateToAllUsers)
					r.Get("/{id}", user_handler.GetUser)
					r.Put("/{id}", user_handler.UpdateUser)
					r.Get("/{id}/channels", user_handler.ListUserChannels)
					r.Put("/{id}/channels", user_handler.ReplaceUserChannels)
					r.Post("/{id}/topup", user_handler.TopUp)
					r.Post("/{id}/reset-password", user_handler.ResetPassword)
					r.Delete("/{id}", user_handler.DeleteUser)
				})

				r.Route("/invites", func(r chi.Router) {
					r.Get("/", user_handler.ListInviteCodes)
					r.Post("/", user_handler.CreateInviteCode)
					r.Get("/settings", user_handler.GetInviteSettings)
					r.Put("/settings", user_handler.UpdateInviteSettings)
				})

				// Channel management
				r.Route("/channels", func(r chi.Router) {
					r.Get("/", channel_handler.ListChannels)
					r.Post("/", channel_handler.CreateChannel)
					r.Get("/{id}", channel_handler.GetChannel)
					r.Put("/{id}", channel_handler.UpdateChannel)
					r.Delete("/{id}", channel_handler.DeleteChannel)
				})

				// Model management
				r.Route("/models", func(r chi.Router) {
					r.Get("/", model_handler.ListModels)
					r.Post("/", model_handler.CreateModel)
					r.Put("/{id}", model_handler.UpdateModel)
					r.Post("/{id}/toggle", model_handler.ToggleModel)
					r.Delete("/{id}", model_handler.DeleteModel)
				})

				// Redeem code management
				r.Route("/redeem-codes", func(r chi.Router) {
					r.Get("/", billing_handler.ListCodes)
					r.Post("/", billing_handler.GenerateCodes)
					r.Delete("/{id}", billing_handler.DeleteCode)
				})

				// Usage statistics
				r.Route("/usage", func(r chi.Router) {
					r.Get("/overview", usage_handler.GetGlobalOverview)
					r.Get("/daily", usage_handler.GetDailyStats)
					r.Get("/top-models", usage_handler.GetTopModels)
					r.Get("/top-users", usage_handler.GetTopUsers)
					r.Get("/logs", usage_handler.GetAllLogs)
				})
			})
		})
	})

	// OpenAI compatible relay routes (API Key auth)
	r.Route("/v1", func(r chi.Router) {
		r.Use(middleware.APIKeyAuth(apikey_repository, rdb))
		r.Post("/chat/completions", relay_handler.ChatCompletion)
		r.Get("/models", relay_handler.GetModels)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	serveWebUI(r)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on %s", addr)

	server := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func serveWebUI(r chi.Router) {
	if !serveWebEnabled() {
		log.Println("Web UI static serving disabled by SERVE_WEB")
		return
	}

	staticDir := resolveWebDistDir()
	if staticDir == "" {
		log.Println("Web dist directory not found, Web UI static serving disabled")
		return
	}
	log.Printf("Serving Web UI from %s", staticDir)

	fileServer := http.FileServer(http.Dir(staticDir))
	r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
		cleanPath := filepath.Clean("/" + req.URL.Path)
		relPath := strings.TrimPrefix(cleanPath, "/")
		if relPath == "" || relPath == "." {
			relPath = "index.html"
		}

		target := filepath.Join(staticDir, relPath)
		if info, err := os.Stat(target); err == nil && !info.IsDir() {
			req.URL.Path = "/" + relPath
			fileServer.ServeHTTP(w, req)
			return
		}

		if filepath.Ext(relPath) != "" {
			http.NotFound(w, req)
			return
		}

		req.URL.Path = "/index.html"
		fileServer.ServeHTTP(w, req)
	})
}

func serveWebEnabled() bool {
	value := strings.TrimSpace(os.Getenv("SERVE_WEB"))
	if value == "" {
		return true
	}
	return strings.EqualFold(value, "true")
}

func resolveWebDistDir() string {
	candidates := []string{
		os.Getenv("WEB_DIST_DIR"),
		"web/dist",
	}
	for _, dir := range candidates {
		if dir == "" {
			continue
		}
		if info, err := os.Stat(filepath.Join(dir, "index.html")); err == nil && !info.IsDir() {
			return dir
		}
	}
	return ""
}
