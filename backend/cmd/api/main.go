package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"meetings-editor/config"
	"meetings-editor/internal/docx"
	repoMeeting "meetings-editor/internal/repository/postgres/meeting"
	repoPerson "meetings-editor/internal/repository/postgres/person"
	svcMeeting "meetings-editor/internal/service/meeting"
	svcPerson "meetings-editor/internal/service/person"
	"meetings-editor/internal/transport/http/handler"
	"meetings-editor/internal/transport/http/middleware"
	"meetings-editor/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("config:", err)
	}

	appLog := logger.New(cfg.Env)

	// Generate a random JWT signing secret (invalidated on restart — acceptable for 1h tokens)
	jwtSecret := make([]byte, 32)
	if _, err := rand.Read(jwtSecret); err != nil {
		log.Fatal("failed to generate JWT secret:", err)
	}

	// DB connection pool
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DBDSN)
	if err != nil {
		appLog.Error(ctx, "failed to connect to database", zap.Error(err))
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		appLog.Error(ctx, "database ping failed", zap.Error(err))
		os.Exit(1)
	}

	// Repositories
	personRepo := repoPerson.New(pool)
	meetingRepo := repoMeeting.New(pool)

	// Services
	personSvc := svcPerson.New(personRepo)
	meetingSvc := svcMeeting.New(meetingRepo, personRepo)

	// Export
	exportGen := docx.New()

	// Handlers
	ph := handler.NewPersonHandler(personSvc)
	mh := handler.NewMeetingHandler(meetingSvc, exportGen)
	ah := handler.NewAuthHandler(jwtSecret, cfg.AdminPassword)

	// Protected routes — require JWT cookie or API key
	protected := http.NewServeMux()

	protected.HandleFunc("GET /people", ph.List)
	protected.HandleFunc("POST /people", ph.Create)
	protected.HandleFunc("POST /people/sort", ph.Sort)
	protected.HandleFunc("GET /people/{id}", ph.GetByID)
	protected.HandleFunc("PATCH /people/{id}", ph.Update)
	protected.HandleFunc("DELETE /people/{id}", ph.Delete)

	protected.HandleFunc("GET /meetings", mh.List)
	protected.HandleFunc("POST /meetings", mh.Create)
	protected.HandleFunc("GET /meetings/{id}", mh.GetByID)
	protected.HandleFunc("GET /meetings/{id}/meta", mh.GetMeta)
	protected.HandleFunc("GET /meetings/{id}/people", mh.GetPeople)
	protected.HandleFunc("GET /meetings/{id}/agenda-items", mh.GetAgendaItems)
	protected.HandleFunc("PATCH /meetings/{id}", mh.Update)
	protected.HandleFunc("DELETE /meetings/{id}", mh.Delete)
	protected.HandleFunc("PUT /meetings/{id}/chairperson", mh.SetChairperson)
	protected.HandleFunc("POST /meetings/{id}/people", mh.AddPerson)
	protected.HandleFunc("DELETE /meetings/{id}/people/{pid}", mh.RemovePerson)
	protected.HandleFunc("PUT /meetings/{id}/people/order", mh.ReorderPeople)
	protected.HandleFunc("POST /meetings/{id}/agenda-items", mh.AddAgendaItem)
	protected.HandleFunc("PUT /meetings/{id}/agenda-items/{item_id}", mh.UpdateAgendaItem)
	protected.HandleFunc("DELETE /meetings/{id}/agenda-items/{item_id}", mh.DeleteAgendaItem)
	protected.HandleFunc("POST /meetings/{id}/agenda-items/{item_id}/speakers", mh.AddAgendaItemSpeaker)
	protected.HandleFunc("DELETE /meetings/{id}/agenda-items/{item_id}/speakers/{pid}", mh.RemoveAgendaItemSpeaker)
	protected.HandleFunc("PUT /meetings/{id}/agenda-items/{item_id}/speakers/order", mh.ReorderAgendaItemSpeakers)
	protected.HandleFunc("PUT /meetings/{id}/agenda-items/order", mh.ReorderAgendaItems)
	protected.HandleFunc("GET /meetings/{id}/export/agenda", mh.ExportAgenda)
	protected.HandleFunc("GET /meetings/{id}/export/participants", mh.ExportParticipants)

	// Main router: public routes + auth-guarded catch-all
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("POST /auth/login", ah.Login)
	mux.HandleFunc("POST /auth/logout", ah.Logout)
	mux.Handle("/", middleware.Auth(jwtSecret, cfg.APIKey)(protected))

	// Chain middleware: CORS → Logging → mux
	var httpHandler http.Handler = mux
	httpHandler = middleware.Logging(appLog)(httpHandler)
	httpHandler = middleware.CORS(httpHandler)

	server := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      httpHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		appLog.Info(ctx, "server starting", zap.String("addr", cfg.HTTPAddr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLog.Error(ctx, "server error", zap.Error(err))
			os.Exit(1)
		}
	}()

	<-quit
	appLog.Info(ctx, "shutting down gracefully")

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutCtx); err != nil {
		appLog.Error(ctx, "shutdown error", zap.Error(err))
	}
}
