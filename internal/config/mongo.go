package config

import (
	"errors"

	"github.com/caarlos0/env/v6"
)

type DB struct {
	Database   string
	Collection string
}

type Mongo struct {
	Host         string `env:"MONGO_HOST" envDefault:"localhost"`
	Username     string `env:"MONGO_USER" envDefault:""`
	Password     string `env:"MONGO_PASS" envDefault:""`
	User         DB
	Hooks        DB
	Agent        DB
	Orchestrator DB
}

func BuildMongo(c *Config) error {
	mongo := &Mongo{}

	if err := env.Parse(mongo); err != nil {
		return err
	}

	creds, err := c.getVaultSecrets("kv/data/k8sdeploy/orchestrator/mongodb")
	if err != nil {
		return err
	}

	if creds == nil {
		return errors.New("no mongo password found")
	}

	kvs, err := ParseKVSecrets(creds)
	if err != nil {
		return err
	}
	if len(kvs) == 0 {
		return errors.New("no mongo details found")
	}

	kvStrings := KVStrings(kvs)
	mongo.Password = kvStrings["password"]
	mongo.Username = kvStrings["username"]
	mongo.Host = kvStrings["hostname"]

	mongo.User.Database = kvStrings["user_db"]
	mongo.User.Collection = kvStrings["user_keys_collection"]
	mongo.Hooks.Database = kvStrings["hooks_db"]
	mongo.Hooks.Collection = kvStrings["hooks_keys_collection"]
	mongo.Agent.Database = kvStrings["agent_db"]
	mongo.Agent.Collection = kvStrings["agent_keys_collection"]
	mongo.Orchestrator.Database = kvStrings["orchestrator_db"]
	mongo.Orchestrator.Collection = kvStrings["orchestrator_agent_collection"]

	c.Mongo = *mongo

	return nil
}
