package main

import (
	"context"
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

	// Router (Go 1.22+ method+path patterns)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	mux.HandleFunc("GET /people", ph.List)
	mux.HandleFunc("POST /people", ph.Create)
	mux.HandleFunc("GET /people/{id}", ph.GetByID)
	mux.HandleFunc("PATCH /people/{id}", ph.Update)
	mux.HandleFunc("DELETE /people/{id}", ph.Delete)

	mux.HandleFunc("GET /meetings", mh.List)
	mux.HandleFunc("POST /meetings", mh.Create)
	mux.HandleFunc("GET /meetings/{id}", mh.GetByID)
	mux.HandleFunc("GET /meetings/{id}/meta", mh.GetMeta)
	mux.HandleFunc("GET /meetings/{id}/people", mh.GetPeople)
	mux.HandleFunc("GET /meetings/{id}/agenda-items", mh.GetAgendaItems)
	mux.HandleFunc("PATCH /meetings/{id}", mh.Update)
	mux.HandleFunc("DELETE /meetings/{id}", mh.Delete)
	mux.HandleFunc("PUT /meetings/{id}/chairperson", mh.SetChairperson)
	mux.HandleFunc("POST /meetings/{id}/people", mh.AddPerson)
	mux.HandleFunc("DELETE /meetings/{id}/people/{pid}", mh.RemovePerson)
	mux.HandleFunc("PUT /meetings/{id}/people/order", mh.ReorderPeople)
	mux.HandleFunc("POST /meetings/{id}/agenda-items", mh.AddAgendaItem)
	mux.HandleFunc("PUT /meetings/{id}/agenda-items/{item_id}", mh.UpdateAgendaItem)
	mux.HandleFunc("DELETE /meetings/{id}/agenda-items/{item_id}", mh.DeleteAgendaItem)
	mux.HandleFunc("POST /meetings/{id}/agenda-items/{item_id}/speakers", mh.AddAgendaItemSpeaker)
	mux.HandleFunc("DELETE /meetings/{id}/agenda-items/{item_id}/speakers/{pid}", mh.RemoveAgendaItemSpeaker)
	mux.HandleFunc("PUT /meetings/{id}/agenda-items/{item_id}/speakers/order", mh.ReorderAgendaItemSpeakers)
	mux.HandleFunc("PUT /meetings/{id}/agenda-items/order", mh.ReorderAgendaItems)
	mux.HandleFunc("GET /meetings/{id}/export/agenda", mh.ExportAgenda)
	mux.HandleFunc("GET /meetings/{id}/export/participants", mh.ExportParticipants)

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
