package config

import "github.com/caarlos0/env/v6"

type K8sDeploy struct {
	SocketAddress string `env:"SOCKET_ADDRESS" envDefault:"https://sockets.chewedfeed.com"`
	Key           string `env:"K8SDEPLOY_KEY" envDefault:""`
	Secret        string `env:"K8SDEPLOY_SECRET" envDefault:""`
}

func BuildK8sDeploy(c *Config) error {
	cfg := &K8sDeploy{}

	if err := env.Parse(cfg); err != nil {
		return err
	}
	c.K8sDeploy = *cfg

	return nil
}
