package manager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/webhook/model"
)

// DefaultManager ...
type DefaultManager struct {
}

// NewDefaultManger ...
func NewDefaultManger() *DefaultManager {
	return &DefaultManager{}
}

// Create webhook policy
func (m *DefaultManager) Create(policy *models.WebhookPolicy) (int64, error) {
	t := time.Now()
	policy.CreationTime = t
	policy.UpdateTime = t

	err := policy.ConvertToDBModel()
	if err != nil {
		return 0, err
	}
	return dao.AddWebhookPolicy(policy)
}

// List the webhook policies, returns the total count, policy list and error
func (m *DefaultManager) List(projectID int64) (int64, []*models.WebhookPolicy, error) {
	policies := []*models.WebhookPolicy{}
	total, persisPolicies, err := dao.GetWebhookPolicies(projectID)
	if err != nil {
		return total, nil, err
	}

	for _, policy := range persisPolicies {
		err := policy.ConvertFromDBModel()
		if err != nil {
			return 0, nil, err
		}
		policies = append(policies, policy)
	}

	return total, policies, nil
}

// Get webhook policy with specified ID
func (m *DefaultManager) Get(id int64) (*models.WebhookPolicy, error) {
	policy, err := dao.GetWebhookPolicy(id)
	if err != nil {
		return nil, err
	}
	err = policy.ConvertFromDBModel()
	return policy, err
}

// GetByNameAndProjectID webhook policy by the name and projectID
func (m *DefaultManager) GetByNameAndProjectID(name string, projectID int64) (*models.WebhookPolicy, error) {
	policy, err := dao.GetWebhookPolicyByName(name, projectID)
	if err != nil {
		return nil, err
	}
	err = policy.ConvertFromDBModel()
	return policy, err
}

// Update the specified webhook policy
func (m *DefaultManager) Update(policy *models.WebhookPolicy) error {
	policy.UpdateTime = time.Now()
	err := policy.ConvertToDBModel()
	if err != nil {
		return err
	}
	return dao.UpdateWebhookPolicy(policy)
}

// Delete the specified webhook policy
func (m *DefaultManager) Delete(policyID int64) error {
	return dao.DeleteWebhookPolicy(policyID)
}

// Test the specified webhook policy, just test for network connection without request body
func (m *DefaultManager) Test(policy *models.WebhookPolicy) error {
	p, err := json.Marshal(model.Payload{
		Type: model.EventTypeTestEndpoint,
	})
	if err != nil {
		return err
	}

	for _, target := range policy.Targets {
		switch target.Type {
		case "http":
			return m.policyHTTPTest(target.Address, target.SkipCertVerify, p)
		default:
			return fmt.Errorf("invalid policy target type: %s", target.Type)
		}
	}
	return nil
}

func (m *DefaultManager) policyHTTPTest(address string, skipCertVerify bool, p []byte) error {
	req, err := http.NewRequest(http.MethodPost, address, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{
		Transport: registry.GetHTTPTransport(skipCertVerify),
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 300 {
		return fmt.Errorf("policy test failed with response code %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("read policy test response body failed: %v", err)
	}
	log.Debugf("policy test response code %d, body: %s", resp.StatusCode, string(body))

	return nil
}

// GetRelatedPolices get policies including hook type in project
func (m *DefaultManager) GetRelatedPolices(projectID int64, hookType string) ([]*models.WebhookPolicy, error) {
	_, policies, err := m.List(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook policies with projectID %d: %v", projectID, err)
	}

	var result []*models.WebhookPolicy

	for _, ply := range policies {
		if !ply.Enabled {
			continue
		}
		for _, t := range ply.HookTypes {
			if t != hookType {
				continue
			}
			result = append(result, ply)
		}
	}
	return result, nil
}
