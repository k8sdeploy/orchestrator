package service

import (
	"fmt"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	bugMiddleware "github.com/bugfixes/go-bugfixes/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"
	"github.com/k8sdeploy/orchestrator-service/internal/config"
	"github.com/k8sdeploy/orchestrator-service/internal/orchestrator"
	"github.com/keloran/go-healthcheck"
	"github.com/keloran/go-probe"
	"net/http"
)

type Service struct {
	Config *config.Config
}

func (s *Service) Start() error {
	errChan := make(chan error)
	go s.startHTTP(errChan)

	return <-errChan
}

func (s *Service) checkAPIKey(next http.Handler) http.Handler {
	r := func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(r)
}

func (s *Service) startHTTP(errChan chan error) {
	p := fmt.Sprintf(":%d", s.Config.Local.HTTPPort)
	bugLog.Local().Infof("Starting hooks-service on %s", p)

	r := chi.NewRouter()
	r.Get("/health", healthcheck.HTTP)
	r.Get("/probe", probe.HTTP)

	r.Route("/", func(r chi.Router) {
		r.Use(middleware.RequestID)
		r.Use(bugMiddleware.BugFixes)
		r.Use(httplog.RequestLogger(httplog.NewLogger("hooks-service", httplog.Options{
			JSON: true,
		})))

		if !s.Config.Local.Development {
			r.Use(s.checkAPIKey)
		}

		r.Post("/agent", orchestrator.NewOrchestrator(s.Config).HandleNewAgent)
	})

	if err := http.ListenAndServe(p, r); err != nil {
		errChan <- err
	}
}
