package service

import (
	"fmt"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	bugMiddleware "github.com/bugfixes/go-bugfixes/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"
	kitlog "github.com/go-kit/log"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/kit"
	"github.com/k8sdeploy/orchestrator-service/internal/config"
	"github.com/k8sdeploy/orchestrator-service/internal/orchestrator"
	pb "github.com/k8sdeploy/protos/generated/orchestrator/v1"
	"github.com/keloran/go-healthcheck"
	"github.com/keloran/go-probe"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"net/http"
)

type Service struct {
	Config *config.Config
}

func (s *Service) Start() error {
	errChan := make(chan error)
	go s.startHTTP(errChan)
	go s.startGRPC(errChan)

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
	bugLog.Local().Infof("Starting orchestrator on %s", p)

	r := chi.NewRouter()
	if !s.Config.Local.Development {
		r.Get("/health", healthcheck.HTTP)
		r.Get("/probe", probe.HTTP)
	}

	r.Route("/", func(r chi.Router) {
		r.Use(middleware.RequestID)
		r.Use(bugMiddleware.BugFixes)
		r.Use(httplog.RequestLogger(httplog.NewLogger("orchestrator", httplog.Options{
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

func (s *Service) startGRPC(errChan chan error) {
	kOpts := []kit.Option{
		kit.WithDecider(func(methodFullName string, err error) bool {
			if err != nil {
				return false
			}
			return true
		}),
	}
	opts := []grpc.ServerOption{
		grpc_middleware.WithStreamServerChain(
			kit.StreamServerInterceptor(kitlog.NewNopLogger(), kOpts...),
		),
		grpc_middleware.WithUnaryServerChain(
			kit.UnaryServerInterceptor(kitlog.NewNopLogger(), kOpts...),
		),
	}
	p := fmt.Sprintf(":%d", s.Config.Local.GRPCPort)
	lis, err := net.Listen("tcp", p)
	if err != nil {
		errChan <- err
	}
	gs := grpc.NewServer(opts...)
	reflection.Register(gs)
	pb.RegisterOrchestratorServer(gs, &orchestrator.Server{
		Config: s.Config,
	})
	if err := gs.Serve(lis); err != nil {
		errChan <- err
	}
}
