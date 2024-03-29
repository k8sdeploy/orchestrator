package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/k8sdeploy/orchestrator-service/internal/config"
	pb "github.com/k8sdeploy/protos/generated/orchestrator/v1"
)

type Server struct {
	pb.UnimplementedOrchestratorServer
	Config *config.Config
}

type ChannelDetails struct {
	Channel   string
	EmitToken string
}

type DeployMessage struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	ImageURL  string `json:"image_url"`
}
type DeployBody struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

func (s *Server) Deploy(ctx context.Context, in *pb.DeploymentRequest) (*pb.DeploymentResponse, error) {
	channel, err := s.GetChannel(in.K8SDetails.HookDetails.Id, in.K8SDetails.HookDetails.Key, in.K8SDetails.HookDetails.Secret)
	if err != nil {
		fmt.Printf("Get Channel err: %+v\n", err)
		return &pb.DeploymentResponse{
			Deployed: false,
		}, nil
	}

	imageVersion := "latest"
	if in.K8SDetails.ImageHash != "" {
		imageVersion = in.K8SDetails.ImageHash
		if len(imageVersion) >= 7 {
			imageVersion = fmt.Sprintf("sha-%s", imageVersion[:7])
		} else {
			imageVersion = fmt.Sprintf("sha-%s", imageVersion)
		}
	}
	if in.K8SDetails.ImageTag != "" {
		imageVersion = in.K8SDetails.ImageTag
	}
	imageURL := fmt.Sprintf("%s/%s:%s", "containers.chewed-k8s.net/k8sdeploy", in.K8SDetails.ServiceName, imageVersion)

	deployMessage := DeployMessage{
		Namespace: in.K8SDetails.ServiceNamespace,
		Name:      in.K8SDetails.ServiceName,
		ImageURL:  imageURL,
	}
	dm, err := json.Marshal(deployMessage)
	if err != nil {
		fmt.Printf("Marshal deployMessage err: %+v\n", err)
		return &pb.DeploymentResponse{
			Deployed: false,
		}, nil
	}

	if err := s.sendMessage(DeployBody{
		Title:   "deploy",
		Message: string(dm),
	}, channel); err != nil {
		fmt.Printf("sendMessage err: %+v\n", err)
		return &pb.DeploymentResponse{
			Deployed: false,
		}, nil
	}

	return &pb.DeploymentResponse{
		Deployed: true,
	}, nil
}

func (s *Server) sendMessage(dep DeployBody, channel ChannelDetails) error {
	b, err := json.Marshal(dep)
	if err != nil {
		fmt.Printf("failed to marshal: %+v\n", err)
		return err
	}

	fmt.Printf("deploy message: %+v\nchannel: %+v\nb:%s\n", dep, channel, b)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/message", s.Config.K8sDeploy.SocketAddress), bytes.NewBuffer(b))
	if err != nil {
		fmt.Printf("failed to create request: %+v\n", err)
		return err
	}

	req.Header.Set("X-Gotify-Key", channel.EmitToken)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)

	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Printf("failed to close response body: %+v\n", err)
		}
	}()

	if err != nil {
		fmt.Printf("channel notify error: %+v\n", err)
		return err
	}
	if res.StatusCode != http.StatusOK {
		fmt.Printf("failed result: %+v\n", res)
		return err
	}

	return nil
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
