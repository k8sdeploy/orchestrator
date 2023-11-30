package agent

import (
	"github.com/k8sdeploy/orchestrator-service/internal/config"
	"net/http"
)

type Agent struct {
	Config *config.Config
}

func NewAgent(cfg *config.Config) *Agent {
	return &Agent{
		Config: cfg,
	}
}

func (a *Agent) Router() http.Handler {
	r := http.NewServeMux()

	return r
}
