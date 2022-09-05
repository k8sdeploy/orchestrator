package config

import (
	"errors"

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
	creds, err := c.getVaultSecrets("kv/data/k8sdeploy/api-keys")
	if err != nil {
		return err
	}
	if creds == nil {
		return errors.New("no api keys found")
	}

	kvs, err := ParseKVSecrets(creds)
	if err != nil {
		return err
	}
	if len(kvs) == 0 {
		return errors.New("no api keys parsed")
	}
	for _, kv := range kvs {
		if kv.Key == "orchestrator" {
			c.K8sDeploy.Key = kv.Value
			c.K8sDeploy.KeyService.Key = kv.Value
		}
	}

	return nil
}

func getChannelKeys(c *Config) error {
	creds, err := c.getVaultSecrets("kv/data/k8sdeploy/orchestrator/channel-keys")
	if err != nil {
		return err
	}
	if creds == nil {
		return errors.New("no channel keys found")
	}

	kvs, err := ParseKVSecrets(creds)
	if err != nil {
		return err
	}
	if len(kvs) == 0 {
		return errors.New("no channel keys parsed")
	}
	for _, kv := range kvs {
		if kv.Key == "update" {
			c.K8sDeploy.UpdateChannelKey = kv.Value
		}
		if kv.Key == "createAccount" {
			c.K8sDeploy.CreateAccount = kv.Value
		}
		if kv.Key == "updateChannelID" {
			c.K8sDeploy.UpdateChannelID = kv.Value
		}
	}

	return nil
}
