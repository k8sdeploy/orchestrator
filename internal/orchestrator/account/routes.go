package account

import (
	chi "github.com/go-chi/chi/v5"
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

	r.Get("/", a.HandleGet)
	r.Put("/", a.HandlePut)
	r.Post("/", a.HandlePost)
	r.Delete("/", a.HandleDelete)

	return r
}

func (a *Account) HandleDelete(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (a *Account) HandleGet(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (a *Account) HandlePost(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (a *Account) HandlePut(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
