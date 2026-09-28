// Command api is the Secrets Center HTTP backend.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"github.com/kovalit/secrets-center/backend/internal/appsecrets"
	"github.com/kovalit/secrets-center/backend/internal/audit"
	"github.com/kovalit/secrets-center/backend/internal/auth"
	"github.com/kovalit/secrets-center/backend/internal/backup"
	"github.com/kovalit/secrets-center/backend/internal/companies"
	"github.com/kovalit/secrets-center/backend/internal/config"
	"github.com/kovalit/secrets-center/backend/internal/crypto"
	"github.com/kovalit/secrets-center/backend/internal/db"
	"github.com/kovalit/secrets-center/backend/internal/extension"
	"github.com/kovalit/secrets-center/backend/internal/httpx"
	"github.com/kovalit/secrets-center/backend/internal/mailer"
	"github.com/kovalit/secrets-center/backend/internal/passwords"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	// Apply migrations on startup so the service is ready in one command.
	if err := db.Migrate(ctx, pool); err != nil {
		return err
	}

	enc, err := crypto.NewEncryptor(cfg.MasterKeyBase64)
	if err != nil {
		return err
	}

	store := db.NewStore(pool)
	auditSvc := audit.New(store)
	mail := mailer.New(mailer.Config{
		Host: cfg.SMTPHost, Port: cfg.SMTPPort, User: cfg.SMTPUser, Password: cfg.SMTPPassword,
		InsecureTLS: cfg.SMTPInsecureTLS, From: cfg.EmailFrom,
	})

	authSvc := auth.NewService(store, mail)
	authHandler := auth.NewHandler(authSvc, auditSvc, cfg.CookieDomain, cfg.CookieSecure)
	passwordsHandler := passwords.NewHandler(passwords.New(store, enc), auditSvc)
	companiesHandler := companies.NewHandler(store)
	appSecretsHandler := appsecrets.NewHandler(appsecrets.New(store, enc), auditSvc)
	extensionHandler := extension.NewHandler(extension.New(store, enc), auditSvc)
	backupHandler := backup.NewHandler(backup.New(store), auditSvc)

	r := chi.NewRouter()
	r.Use(httpx.RequestID)
	r.Use(httpx.Logger)
	r.Use(httpx.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID", "X-Export-SHA256"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			httpx.JSON(w, http.StatusServiceUnavailable, map[string]string{"status": "db_unavailable"})
			return
		}
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	authLimiter := httpx.NewRateLimiter(10, time.Minute)

	r.Route("/api", func(api chi.Router) {
		// Auth endpoints: public ones rate-limited, account ones behind a session.
		api.Mount("/auth", authHandler.Routes(authSvc.RequireAuth, authLimiter.Middleware))

		// Extension endpoints: token mgmt behind a session, lookup/reveal behind bearer.
		api.Mount("/extension", extensionHandler.Routes(authSvc.RequireAuth))

		// Protected endpoints (web session cookie).
		api.Group(func(priv chi.Router) {
			priv.Use(authSvc.RequireAuth)
			priv.Get("/password-groups", passwordsHandler.ListGroups)
			priv.Mount("/passwords", passwordsHandler.Routes())
			priv.Mount("/companies", companiesHandler.Routes())
			priv.Mount("/app-projects", appSecretsHandler.ProjectRoutes())
			priv.Mount("/app-environments", appSecretsHandler.EnvironmentRoutes())
			priv.Mount("/app-secrets", appSecretsHandler.SecretRoutes())
			priv.Mount("/backups", backupHandler.Routes())
			priv.Get("/audit-logs", auditLogsHandler(store))
		})
	})

	srv := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Printf("secrets-center backend listening on :%s (env=%s)", cfg.AppPort, cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

// auditLogsHandler returns recent audit entries for the current user.
func auditLogsHandler(store *db.Store) http.HandlerFunc {
	type auditResponse struct {
		ID         string    `json:"id"`
		Action     string    `json:"action"`
		EntityType string    `json:"entity_type"`
		EntityID   *string   `json:"entity_id"`
		IP         *string   `json:"ip"`
		CreatedAt  time.Time `json:"created_at"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.CurrentUserID(r.Context())
		logs, err := store.ListAuditLogs(r.Context(), userID, 100)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		out := make([]auditResponse, 0, len(logs))
		for i := range logs {
			l := &logs[i]
			out = append(out, auditResponse{ID: l.ID, Action: l.Action, EntityType: l.EntityType, EntityID: l.EntityID, IP: l.IP, CreatedAt: l.CreatedAt})
		}
		httpx.JSON(w, http.StatusOK, out)
	}
}
