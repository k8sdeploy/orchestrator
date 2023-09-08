package main

import (
	"fmt"

	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/k8sdeploy/orchestrator-service/internal/config"
	"github.com/k8sdeploy/orchestrator-service/internal/service"
)

var (
	BuildVersion = "dev"
	BuildHash    = "unknown"
	ServiceName  = "base-service"
)

func main() {
	logs.Local().Info(fmt.Sprintf("Starting %s", ServiceName))
	logs.Local().Info(fmt.Sprintf("Version: %s, Hash: %s", BuildVersion, BuildHash))

	cfg, err := config.Build()
	if err != nil {
		_ = logs.Errorf("config: %v", err)
		return
	}

	s := &service.Service{
		Config: cfg,
	}

	if err := s.Start(); err != nil {
		_ = logs.Errorf("start service: %v", err)
		return
	}
}
