package project

import (
	"net/http"

	"github.com/k8sdeploy/orchestrator-service/internal/config"
)

type Project struct {
	Config *config.Config
}

func NewProject(cfg *config.Config) *Project {
	return &Project{
		Config: cfg,
	}
}

func (p *Project) Router() http.Handler {
	r := http.NewServeMux()

	return r
}
