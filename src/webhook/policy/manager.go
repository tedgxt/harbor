// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package policy

import (
	"encoding/json"
	"time"

	"github.com/goharbor/harbor/src/common/dao"
	persist_models "github.com/goharbor/harbor/src/common/models"
	rep_models "github.com/goharbor/harbor/src/replication/models"
	"github.com/goharbor/harbor/src/webhook/models"
	"github.com/goharbor/harbor/src/replication"
)

// Manager defines the method a policy manger should implement
type Manager interface {
	GetPolicies(projectID int64, name string) ([]*models.WebhookPolicy, error)
	GetEnabledPolicies(projectID int64) ([]*models.WebhookPolicy, error)
	GetPolicy(int64) (models.WebhookPolicy, error)
	CreatePolicy(models.WebhookPolicy) (int64, error)
	UpdatePolicy(models.WebhookPolicy) error
	RemovePolicy(int64) error
}

// DefaultManager provides webhook policy CURD capabilities.
type DefaultManager struct{}

// NewDefaultManager is the constructor of DefaultManager.
func NewDefaultManager() *DefaultManager {
	return &DefaultManager{}
}

// GetPolicies returns all the policies in project
func (m *DefaultManager) GetPolicies(projectID int64, name string) ([]*models.WebhookPolicy, error) {
	policies, err := dao.GetWebhookPolicyByName(projectID, name)
	if err != nil {
		return nil, err
	}

	list := []*models.WebhookPolicy{}
	for _, policy := range policies {
		ply, err := convertFromForWP(policy)
		if err != nil {
			return nil, err
		}

		list = append(list, &ply)
	}

	return list, nil
}

// GetEnabledPolicies returns all the enabled policies in project
func (m *DefaultManager) GetEnabledPolicies(projectID int64) ([]*models.WebhookPolicy, error) {
	policies, err := dao.GetEnabledWebhookPolicyByProject(projectID)
	if err != nil {
		return nil, err
	}

	list := []*models.WebhookPolicy{}
	for _, policy := range policies {
		ply, err := convertFromForWP(policy)
		if err != nil {
			return nil, err
		}

		list = append(list, &ply)
	}

	return list, nil
}

// GetPolicy returns the policy with the specified ID
func (m *DefaultManager) GetPolicy(policyID int64) (models.WebhookPolicy, error) {
	policy, err := dao.GetWebhookPolicy(policyID)
	if err != nil {
		return models.WebhookPolicy{}, err
	}

	return convertFromForWP(policy)
}

func convertFromForWP(policy *persist_models.WebhookPolicy) (models.WebhookPolicy, error) {
	if policy == nil {
		return models.WebhookPolicy{}, nil
	}

	ply := models.WebhookPolicy{
		ID:                policy.ID,
		Name:              policy.Name,
		Description:       policy.Description,
		ProjectID:         policy.ProjectID,
		Target:            policy.Target,
		CreationTime:      policy.CreationTime,
		UpdateTime:        policy.UpdateTime,
		Enabled:           policy.Enabled,
	}

	if len(policy.Filters) > 0 {
		filters := []rep_models.Filter{}
		if err := json.Unmarshal([]byte(policy.Filters), &filters); err != nil {
			return models.WebhookPolicy{}, err
		}
		for i := range filters {
			if filters[i].Value == nil && len(filters[i].Pattern) > 0 {
				filters[i].Value = filters[i].Pattern
			}
			// convert the type of Value to int64 as the default type of
			// json Unmarshal for number is float64
			if filters[i].Kind == replication.FilterItemKindLabel {
				filters[i].Value = int64(filters[i].Value.(float64))
			}
		}
		ply.Filters = filters
	}

	if len(policy.HookTypes) > 0 {
		hookTypes := []string{}
		if err := json.Unmarshal([]byte(policy.HookTypes), &hookTypes); err != nil {
			return models.WebhookPolicy{}, err
		}
		ply.HookTypes = hookTypes
	}

	return ply, nil
}

func convertToForWP(policy models.WebhookPolicy) (*persist_models.WebhookPolicy, error) {
	ply := &persist_models.WebhookPolicy{
		ID:                policy.ID,
		Name:              policy.Name,
		Description:       policy.Description,
		ProjectID:         policy.ProjectID,
		Target:            policy.Target,
		CreationTime:      policy.CreationTime,
		UpdateTime:        policy.UpdateTime,
		Enabled:           policy.Enabled,
	}

	if len(policy.HookTypes) > 0 {
		hookTypes, err := json.Marshal(policy.HookTypes)
		if err != nil {
			return nil, err
		}
		ply.HookTypes = string(hookTypes)
	}

	if len(policy.Filters) > 0 {
		filters, err := json.Marshal(policy.Filters)
		if err != nil {
			return nil, err
		}
		ply.Filters = string(filters)
	}

	return ply, nil
}

// CreatePolicy creates a new policy with the provided data;
// If creating failed, error will be returned;
// If creating succeed, ID of the new created policy will be returned.
func (m *DefaultManager) CreatePolicy(policy models.WebhookPolicy) (int64, error) {
	now := time.Now()
	policy.CreationTime = now
	policy.UpdateTime = now
	ply, err := convertToForWP(policy)
	if err != nil {
		return 0, err
	}
	return dao.AddWebhookPolicy(*ply)
}

// UpdatePolicy updates the policy;
// If updating failed, error will be returned.
func (m *DefaultManager) UpdatePolicy(policy models.WebhookPolicy) error {
	policy.UpdateTime = time.Now()
	ply, err := convertToForWP(policy)
	if err != nil {
		return err
	}
	return dao.UpdateWebhookPolicy(ply)
}

// RemovePolicy removes the specified policy;
// If removing failed, error will be returned.
func (m *DefaultManager) RemovePolicy(policyID int64) error {
	return dao.DeleteWebhookPolicy(policyID)
}

