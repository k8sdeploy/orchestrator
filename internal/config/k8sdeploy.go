package config

import (
	"github.com/bugfixes/go-bugfixes/logs"
	vault_helper "github.com/keloran/vault-helper"

	"github.com/caarlos0/env/v6"
)

type K8sDeploy struct {
	SocketAddress    string `env:"SOCKET_ADDRESS" envDefault:"https://sockets.chewedfeed.com"`
	Key              string `env:"K8SDEPLOY_KEY" envDefault:""`
	Secret           string `env:"K8SDEPLOY_SECRET" envDefault:""`
	KeyService       KeyService
	UpdateChannelKey string `env:"UPDATE_CHANNEL_KEY" envDefault:""`
	UpdateChannelID  string `env:"UPDATE_CHANNEL_ID" envDefault:""`
	CreateAccount    string `env:"CREATE_ACCOUNT" envDefault:""`
}

type KeyService struct {
	Address string `env:"KEY_SERVICE_ADDRESS" envDefault:"key-service.k8sdeploy:8001"`
	Key     string `env:"KEY_SERVICE_KEY" envDefault:""`
}

func BuildK8sDeploy(c *Config) error {
	cfg := &K8sDeploy{}

	if err := env.Parse(cfg); err != nil {
		return err
	}
	c.K8sDeploy = *cfg

	if err := getAPIKeys(c); err != nil {
		return err
	}

	if err := getChannelKeys(c); err != nil {
		return err
	}

	return nil
}

func getAPIKeys(c *Config) error {
	vh := vault_helper.NewVault(c.Config.Vault.Address, c.Config.Vault.Token)
	if err := vh.GetSecrets("kv/data/k8sdeploy/api-keys"); err != nil {
		return logs.Errorf("get api-keys: %v", err)
	}
	keyService, err := vh.GetSecret("key-service")
	if err != nil {
		return logs.Errorf("get key-service: %v", err)
	}
	orcKey, err := vh.GetSecret("orchestrator")
	if err != nil {
		return logs.Errorf("get orchestrator: %v", err)
	}

	c.K8sDeploy.Key = orcKey
	c.K8sDeploy.KeyService.Key = keyService

	return nil
}

func getChannelKeys(c *Config) error {
	vh := vault_helper.NewVault(c.Config.Vault.Address, c.Config.Vault.Token)
	if err := vh.GetSecrets("kv/data/k8sdeploy/orchestrator/channel-keys"); err != nil {
		return err
	}
	updateChannel, err := vh.GetSecret("update")
	if err != nil {
		return logs.Errorf("get key-service: %v", err)
	}
	createChannel, err := vh.GetSecret("createAccount")
	if err != nil {
		return logs.Errorf("get orchestrator: %v", err)
	}
	updateChannel, err = vh.GetSecret("updateChannelID")
	if err != nil {
		return logs.Errorf("get orchestrator: %v", err)
	}

	c.K8sDeploy.UpdateChannelKey = updateChannel
	c.K8sDeploy.CreateAccount = createChannel
	c.K8sDeploy.UpdateChannelID = updateChannel

	return nil
}
