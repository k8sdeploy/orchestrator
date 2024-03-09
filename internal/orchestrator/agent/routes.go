package agent

import (
	"net/http"

	"github.com/k8sdeploy/orchestrator-service/internal/config"
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
