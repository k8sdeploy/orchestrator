package orchestrator

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/k8sdeploy/orchestrator-service/internal/config"
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
		return nil, logs.Errorf("failed to connect to mongo: %+v", err)
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
			_ = logs.Error(err)
		}
	}()

	var agentData AgentData
	err = client.
		Database(m.Config.Mongo.Database).
		Collection(m.Config.Collections["agents"]).
		FindOne(m.CTX, map[string]string{
			"company_id":   companyID,
			"agent_key":    key,
			"agent_secret": secret,
		}).
		Decode(&agentData)
	if err != nil {
		return nil, logs.Errorf("failed to get agent details: %+v", err)
	}

	return &agentData, nil
}

func (m *Mongo) UpdateAgentChannel(companyID, channelID, channelKey string) error {
	client, err := m.getConnection()
	if err != nil {
		return err
	}
	defer func() {
		if err := client.Disconnect(m.CTX); err != nil {
			_ = logs.Error(err)
		}
	}()

	_, err = client.
		Database(m.Config.Mongo.Database).
		Collection(m.Config.Collections["agents"]).
		UpdateOne(m.CTX,
			bson.D{{
				Key:   "company_id",
				Value: companyID,
			}},
			bson.D{{
				Key: "$set",
				Value: bson.D{{
					Key:   "channel_id",
					Value: channelID,
				}, {
					Key:   "channel_key",
					Value: channelKey,
				}},
			}},
		)
	if err != nil {
		return logs.Errorf("failed to update agent channel: %+v", err)
	}

	return nil
}
