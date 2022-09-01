package orchestrator

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/k8sdeploy/orchestrator-service/internal/config"
)

type Orchestrator struct {
	Config *config.Config
}

func NewOrchestrator(cfg *config.Config) *Orchestrator {
	return &Orchestrator{
		Config: cfg,
	}
}

func (o *Orchestrator) HandleNewAgent(w http.ResponseWriter, r *http.Request) {
	var i interface{}

	if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Printf(" %+v\n", i)
}
