package orchestrator

import (
	"context"
	"github.com/k8sdeploy/orchestrator-service/internal/config"
	pb "github.com/k8sdeploy/protos/generated/orchestrator/v1"
)

type Server struct {
	pb.UnimplementedOrchestratorServer
	Config *config.Config
}

func (s *Server) Deploy(ctx context.Context, in *pb.DeploymentRequest) (*pb.DeploymentResponse, error) {
	return nil, nil
}
