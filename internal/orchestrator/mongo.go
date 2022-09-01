package orchestrator

import (
	"context"
	"fmt"

	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/k8sdeploy/orchestrator-service/internal/config"
	"github.com/mrz1836/go-sanitize"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	Config *config.Config
	CTX    context.Context
}

func NewMongo(c *config.Config) *Mongo {
	return &Mongo{
		Config: c,
		CTX:    context.Background(),
	}
}

func (m *Mongo) getConnection() (*mongo.Client, error) {
	client, err := mongo.Connect(
		m.CTX,
		options.Client().ApplyURI(fmt.Sprintf(
			"mongodb+srv://%s:%s@%s",
			m.Config.Mongo.Username,
			m.Config.Mongo.Password,
			m.Config.Mongo.Host)),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}

type AgentData struct {
	CompanyID   string `json:"company_id" bson:"company_id"`
	ChannelID   string `json:"channel_id" bson:"channel_id"`
	AgentID     string `json:"agent_id" bson:"agent_id"`
	AgentKey    string `json:"agent_key" bson:"agent_key"`
	AgentSecret string `json:"agent_secret" bson:"agent_secret"`
	ChannelKey  string `json:"channel_key" bson:"channel_key"`
	HooksKey    string `json:"hooks_key" bson:"hooks_key"`
	HooksSecret string `json:"hooks_secret" bson:"hooks_secret"`
}

func (m *Mongo) GetAgentDetails(companyID, key, secret string) (*AgentData, error) {
	client, err := m.getConnection()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := client.Disconnect(m.CTX); err != nil {
			bugLog.Info(err)
		}
	}()

	var agentData AgentData
	err = client.
		Database(m.Config.Mongo.Orchestrator.Database).
		Collection(m.Config.Orchestrator.Collection).
		FindOne(m.CTX, map[string]string{
			"company_id":   sanitize.AlphaNumeric(companyID, false),
			"hooks_key":    sanitize.AlphaNumeric(key, false),
			"hooks_secret": sanitize.AlphaNumeric(secret, false)}).
		Decode(&agentData)
	if err != nil {
		return nil, err
	}

	return &agentData, nil
}
