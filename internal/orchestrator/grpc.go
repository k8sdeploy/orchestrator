package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/k8sdeploy/orchestrator-service/internal/config"
	pb "github.com/k8sdeploy/protos/generated/orchestrator/v1"
	"net/http"
)

type Server struct {
	pb.UnimplementedOrchestratorServer
	Config *config.Config
}

type ChannelDetails struct {
	Channel   string
	EmitToken string
}

func (s *Server) Deploy(ctx context.Context, in *pb.DeploymentRequest) (*pb.DeploymentResponse, error) {
	channel, err := s.GetChannel(in.K8SDetails.HookDetails.Id, in.K8SDetails.HookDetails.Key, in.K8SDetails.HookDetails.Secret)
	if err != nil {
		fmt.Printf("Get Channel err: %+v\n", err)
		return &pb.DeploymentResponse{
			Deployed: false,
		}, nil
	}

	type DeployMessage struct {
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
		ImageURL  string `json:"image_url"`
	}
	type DeployBody struct {
		Title         string `json:"title"`
		DeployMessage `json:"message"`
	}

	imageVersion := "latest"
	if in.K8SDetails.ImageTag != "" {
		imageVersion = in.K8SDetails.ImageTag
	}
	if in.K8SDetails.ImageHash != "" {
		imageVersion = in.K8SDetails.ImageHash
	}
	imageURL := fmt.Sprintf("%s@%s", "containers.chewedfeed.com/k8sdeploy/hooks-service", imageVersion)

	dep := DeployBody{
		Title: "deploy",
		DeployMessage: DeployMessage{
			Namespace: in.K8SDetails.ServiceNamespace,
			Name:      in.K8SDetails.ServiceName,
			ImageURL:  imageURL,
		},
	}
	b, err := json.Marshal(dep)
	if err != nil {
		fmt.Printf("failed to marshal: %+v\n", err)
		return &pb.DeploymentResponse{
			Deployed: false,
		}, nil
	}

	fmt.Printf("deploy message: %+v\nchannel: %+v\nb:%s\n", dep, channel, b)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/message", s.Config.K8sDeploy.SocketAddress), bytes.NewBuffer(b))
	req.Header.Set("X-Gotify-Key", channel.EmitToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("channel notify error: %+v\n", err)
		return &pb.DeploymentResponse{
			Deployed: false,
		}, nil
	}
	if res.StatusCode != http.StatusOK {
		fmt.Printf("failed result: %+v\n", res)
		return &pb.DeploymentResponse{
			Deployed: false,
		}, nil
	}

	return &pb.DeploymentResponse{
		Deployed: true,
	}, nil
}

func (s *Server) GetChannel(id, key, secret string) (ChannelDetails, error) {
	channelDetails, err := NewMongo(s.Config).GetAgentDetails(id, key, secret)
	if err != nil {
		return ChannelDetails{}, err
	}

	return ChannelDetails{
		Channel:   channelDetails.ChannelKey,
		EmitToken: channelDetails.ChannelKey,
	}, nil
}
