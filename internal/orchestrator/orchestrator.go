package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/bugfixes/go-bugfixes/logs"

	keybuf "github.com/k8sdeploy/protos/generated/key/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

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

type AgentRequest struct {
	CompanyID string `json:"company_id"`
	Key       string `json:"key"`
	Secret    string `json:"secret"`
}

type AgentChannelDetails struct {
	Token   string `json:"token"`
	Channel string `json:"channel"`
}

type AgentResponse struct {
	Update AgentChannelDetails `json:"update"`
	Event  AgentChannelDetails `json:"event"`
}

func (o *Orchestrator) HandleNewAgent(w http.ResponseWriter, r *http.Request) {
	var ab AgentRequest

	if err := json.NewDecoder(r.Body).Decode(&ab); err != nil {
		logs.Infof("failed to decode request body: %+v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !o.Config.Development {
		validKeys, err := o.validateAgentKeys(ab)
		if err != nil {
			logs.Infof("failed to validate agent keys: %+v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !validKeys {
			logs.Infof("invalid agent keys: %+v\n", ab)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	updateDetails, err := o.GetUpdateDetails(ab)
	if err != nil {
		logs.Infof("failed to get update details: %+v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	eventDetails, err := o.GetEventDetails(ab, updateDetails)
	if err != nil {
		logs.Infof("failed to get event details: %+v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if updateDetails == (AgentChannelDetails{}) && eventDetails == (AgentChannelDetails{}) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := json.NewEncoder(w).Encode(AgentResponse{
		Update: updateDetails,
		Event:  eventDetails,
	}); err != nil {
		logs.Infof("failed to encode response: %+v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	logs.Infof(" %+v\n", ab)
}

func (o *Orchestrator) validateAgentKeys(ab AgentRequest) (bool, error) {
	if o.Config.Development {
		return true, nil
	}

	conn, err := grpc.Dial(o.Config.K8sDeploy.KeyService.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return false, logs.Errorf("failed to dial key service: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			_ = logs.Errorf("failed to close connection: %v", err)
		}
	}()

	k := keybuf.ValidateSystemKeyRequest{
		ServiceKey: o.Config.K8sDeploy.KeyService.Key,
		CompanyId:  ab.CompanyID,
		Key:        ab.Key,
		Secret:     ab.Secret,
	}

	c := keybuf.NewKeyServiceClient(conn)
	resp, err := c.ValidateAgentKey(context.Background(), &k)
	if err != nil {
		return false, logs.Errorf("validateKey failed to validate key: %v", err)
	}

	if resp.Status != nil {
		return false, logs.Errorf("unknown status: %v", *resp.Status)
	}

	if resp.Valid {
		return true, nil
	}

	return false, nil
}

func (o *Orchestrator) GetUpdateDetails(request AgentRequest) (AgentChannelDetails, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/client", o.Config.K8sDeploy.SocketAddress), nil)
	if err != nil {
		return AgentChannelDetails{}, logs.Errorf("failed to create request: %+v", err)
	}

	//req.Header.Set("Authorization", fmt.Sprintf("Basic %s", o.Config.K8sDeploy.CreateAccount))
	req.Header.Set("X-Gotify-Key", o.Config.K8sDeploy.CreateAccount)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return AgentChannelDetails{}, logs.Errorf("failed to get client: %+v", err)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			_ = logs.Errorf("failed to close response body: %+v\n", err)
		}
	}()

	type client struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Token string `json:"token"`
	}
	var cs []client

	if err := json.NewDecoder(res.Body).Decode(&cs); err != nil {
		return AgentChannelDetails{}, logs.Errorf("failed to decode response body: %+v", err)
	}

	if len(cs) != 0 {
		for _, c := range cs {
			if c.Name == request.CompanyID {
				return AgentChannelDetails{
					Token:   c.Token,
					Channel: o.Config.K8sDeploy.UpdateChannelID,
				}, nil
			}
		}
	}

	return AgentChannelDetails{}, nil
}

func (o *Orchestrator) GetEventDetails(request AgentRequest, details AgentChannelDetails) (AgentChannelDetails, error) {
	agent, err := NewMongo(o.Config).GetAgentDetails(request.CompanyID, request.Key, request.Secret)
	if err != nil {
		return AgentChannelDetails{}, logs.Errorf("failed to get agent details: %+v", err)
	}
	if agent.ChannelID == "" {
		return AgentChannelDetails{}, nil
	}

	return AgentChannelDetails{
		Token:   details.Token,
		Channel: agent.ChannelID,
	}, nil
}

func (o *Orchestrator) createUpdateDetails(request AgentRequest) (AgentChannelDetails, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/client", o.Config.K8sDeploy.SocketAddress), strings.NewReader(
		fmt.Sprintf(`{"name": "%s"}`, request.CompanyID)))
	if err != nil {
		return AgentChannelDetails{}, logs.Errorf("failed to create request: %+v", err)
	}
	//req.Header.Set("Authorization", fmt.Sprintf("Basic %s", o.Config.K8sDeploy.CreateAccount))
	req.Header.Set("X-Gotify-Key", o.Config.K8sDeploy.CreateAccount)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return AgentChannelDetails{}, logs.Errorf("failed to create client: %+v", err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			_ = logs.Errorf("failed to close response body: %+v\n", err)
		}
	}()
	type createResponse struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Token string `json:"token"`
	}
	var cr createResponse
	if err := json.NewDecoder(res.Body).Decode(&cr); err != nil {
		return AgentChannelDetails{}, logs.Errorf("failed to decode response body: %+v", err)
	}

	return AgentChannelDetails{
		Token:   cr.Token,
		Channel: o.Config.K8sDeploy.UpdateChannelID,
	}, nil
}

func (o *Orchestrator) HandleNewAgentAccount(w http.ResponseWriter, r *http.Request) {
	var ab AgentRequest
	if err := json.NewDecoder(r.Body).Decode(&ab); err != nil {
		logs.Infof("failed to decode request body: %+v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if ab.CompanyID == "" || ab.Key == "" || ab.Secret == "" {
		logs.Infof("invalid request body: %+v\n", ab)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ud, err := o.createUpdateDetails(ab)
	if err != nil {
		logs.Infof("failed to create update details: %+v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	acd, err := o.createEventDetails(ab, AgentChannelDetails{
		Token: ud.Token,
	})
	if err != nil {
		logs.Infof("failed to create event details: %+v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logs.Infof("created event details: %+v\n", acd)
}

func (o *Orchestrator) createEventDetails(request AgentRequest, details AgentChannelDetails) (AgentChannelDetails, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/application", o.Config.K8sDeploy.SocketAddress), strings.NewReader(
		fmt.Sprintf(`{"name": "%s"}`, request.CompanyID)))
	if err != nil {
		return AgentChannelDetails{}, logs.Errorf("failed to create request: %+v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", o.Config.K8sDeploy.CreateAccount))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return AgentChannelDetails{}, logs.Errorf("failed to create application: %+v", err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			_ = logs.Errorf("failed to close response body: %+v", err)
		}
	}()
	type createResponse struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Token string `json:"token"`
	}
	var cr createResponse
	if err := json.NewDecoder(res.Body).Decode(&cr); err != nil {
		return AgentChannelDetails{}, logs.Errorf("failed to decode response body: %+v", err)
	}

	id := fmt.Sprintf("%d", cr.ID)
	if err := NewMongo(o.Config).UpdateAgentChannel(request.CompanyID, id, cr.Token); err != nil {
		return AgentChannelDetails{}, logs.Errorf("failed to update agent channel details: %+v", err)
	}

	return AgentChannelDetails{
		Token:   details.Token,
		Channel: id,
	}, nil
}
