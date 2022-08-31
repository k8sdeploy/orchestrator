package orchestrator

import (
	"context"
	"fmt"
	"github.com/k8sdeploy/orchestrator-service/internal/config"
	pb "github.com/k8sdeploy/protos/generated/orchestrator/v1"
)

type Server struct {
	pb.UnimplementedOrchestratorServer
	Config *config.Config
}

func (s *Server) Deploy(ctx context.Context, in *pb.DeploymentRequest) (*pb.DeploymentResponse, error) {
	fmt.Printf("Received: %+v\n", in)

	return &pb.DeploymentResponse{
		Deployed: true,
	}, nil
}
