package account

import (
	"github.com/go-chi/chi/v5"
	"github.com/k8sdeploy/orchestrator-service/internal/config"
	"net/http"
)

type Account struct {
	Config *config.Config
}

func NewAccount(cfg *config.Config) *Account {
	return &Account{
		Config: cfg,
	}
}

func (a *Account) Router() http.Handler {
	r := chi.NewRouter()

	return r
}
