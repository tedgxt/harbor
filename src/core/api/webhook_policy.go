// Copyright 2018 Project Harbor Authors
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

package api

import (
	"fmt"

	"net/http"
	"strconv"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	api_models "github.com/goharbor/harbor/src/core/api/models"
	"github.com/goharbor/harbor/src/webhook/controller"
	"github.com/goharbor/harbor/src/replication"
	webhook_models "github.com/goharbor/harbor/src/webhook/models"
	"github.com/goharbor/harbor/src/webhook"
	"errors"
)

// WebhookPolicyAPI handles /api/policies/webhook/:id([0-9]+) /api/policies/webhook
type WebhookPolicyAPI struct {
	BaseController
}

// Prepare ...
func (wpa *WebhookPolicyAPI) Prepare() {
	wpa.BaseController.Prepare()
	if !wpa.SecurityCtx.IsAuthenticated() {
		wpa.HandleUnauthorized()
		return
	}
}

// Get ...
func (wpa *WebhookPolicyAPI) Get() {
	id := wpa.GetIDFromURL()
	policy, err := controller.PolicyManager.GetPolicy(id)
	if err != nil {
		log.Errorf("failed to get webhook policy %d: %v", id, err)
		wpa.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if policy.ID == 0 {
		wpa.HandleNotFound(fmt.Sprintf("webhook policy %d not found", id))
		return
	}

	project, err := wpa.doAuth(policy.ProjectID)
	if err != nil {
		return
	}

	ply, err := convertToAPIModel(project, &policy)
	if err != nil {
		wpa.ParseAndHandleError(fmt.Sprintf("failed to convert from webhook policy"), err)
		return
	}

	wpa.Data["json"] = ply
	wpa.ServeJSON()
}

// List ...
func (wpa *WebhookPolicyAPI) List() {
	name := wpa.GetString("name")
	projectIDStr := wpa.GetString("project_id")
	var projectID int64
	if len(projectIDStr) > 0 {
		var err error
		projectID, err = strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil || projectID <= 0 {
			wpa.HandleBadRequest(fmt.Sprintf("invalid project ID: %s", projectIDStr))
			return
		}
	}

	project, err := wpa.doAuth(projectID)
	if err != nil {
		return
	}

	result, err := controller.PolicyManager.GetPolicies(projectID, name)
	if err != nil {
		log.Errorf("failed to get policies: %v, projectID: %d, name: %s", err, projectID, name)
		wpa.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	policies := []*api_models.WebhookPolicy{}
	if result != nil {
		for _, policy := range result {
			ply, err := convertToAPIModel(project, policy)
			if err != nil {
				wpa.ParseAndHandleError(fmt.Sprintf("failed to convert from webhook policy"), err)
				return
			}
			policies = append(policies, ply)
		}
	}

	wpa.Data["json"] = policies
	wpa.ServeJSON()
}

// Post creates a webhook policy
func (wpa *WebhookPolicyAPI) Post() {
	policy := &api_models.WebhookPolicy{}
	wpa.DecodeJSONReqAndValidate(policy)

	project, err := wpa.doAuth(policy.Project.ProjectID)
	if err != nil {
		return
	}

	// check the name
	exist, err := isExist(project.ProjectID, policy.Name)
	if err != nil {
		wpa.HandleInternalServerError(fmt.Sprintf("failed to check the existence of policy %s: %v", policy.Name, err))
		return
	}

	if exist {
		wpa.HandleConflict(fmt.Sprintf("name %s is already used", policy.Name))
		return
	}

	// check the existence of labels
	for _, filter := range policy.Filters {
		if filter.Kind == replication.FilterItemKindLabel {
			labelID := filter.Value.(int64)
			label, err := dao.GetLabel(labelID)
			if err != nil {
				wpa.HandleInternalServerError(fmt.Sprintf("failed to get label %d: %v", labelID, err))
				return
			}
			if label == nil || label.Deleted {
				wpa.HandleNotFound(fmt.Sprintf("label %d not found", labelID))
				return
			}
		}
	}

	// check hook type
	for _, hookType := range policy.HookTypes {
		typeDefined := false
		for _, definedType := range webhook.HookTypes {
			if definedType == hookType {
				typeDefined = true
				break
			}
		}
		if !typeDefined {
			wpa.HandleBadRequest(fmt.Sprintf("hook type %s not supported", hookType))
			return
		}
	}

	id, err := controller.PolicyManager.CreatePolicy(convertToWebhookPolicy(policy))
	if err != nil {
		wpa.HandleInternalServerError(fmt.Sprintf("failed to create policy: %v", err))
		return
	}

	wpa.Redirect(http.StatusCreated, strconv.FormatInt(id, 10))
}

func isExist(projectID int64, name string) (bool, error) {
	result, err := controller.PolicyManager.GetPolicies(projectID, name)
	if err != nil {
		return false, err
	}

	for _, policy := range result {
		if policy.Name == name {
			return true, nil
		}
	}
	return false, nil
}

// Put updates the webhook policy
func (wpa *WebhookPolicyAPI) Put() {
	id := wpa.GetIDFromURL()

	originalPolicy, err := controller.PolicyManager.GetPolicy(id)
	if err != nil {
		log.Errorf("failed to get policy %d: %v", id, err)
		wpa.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if originalPolicy.ID == 0 {
		wpa.HandleNotFound(fmt.Sprintf("policy %d not found", id))
		return
	}

	if _, err := wpa.doAuth(originalPolicy.ProjectID); err != nil {
		return
	}

	policy := &api_models.WebhookPolicy{}
	wpa.DecodeJSONReqAndValidate(policy)

	policy.ID = id

	// check projectID, projectID shouldn't be changed
	if originalPolicy.ProjectID != policy.Project.ProjectID {
		wpa.HandleBadRequest("project shouldn't be changed")
		return
	}

	// check the name
	if policy.Name != originalPolicy.Name {
		exist, err := exist(policy.Name)
		if err != nil {
			wpa.HandleInternalServerError(fmt.Sprintf("failed to check the existence of policy %s: %v", policy.Name, err))
			return
		}

		if exist {
			wpa.HandleConflict(fmt.Sprintf("name %s is already used", policy.Name))
			return
		}
	}

	// check hook type
	for _, hookType := range policy.HookTypes {
		typeDefined := false
		for _, definedType := range webhook.HookTypes {
			if definedType == hookType {
				typeDefined = true
				break
			}
		}
		if !typeDefined {
			wpa.HandleBadRequest(fmt.Sprintf("hook type %s not supported", hookType))
			return
		}
	}

	// check the existence of labels
	for _, filter := range policy.Filters {
		if filter.Kind == replication.FilterItemKindLabel {
			labelID := filter.Value.(int64)
			label, err := dao.GetLabel(labelID)
			if err != nil {
				wpa.HandleInternalServerError(fmt.Sprintf("failed to get label %d: %v", labelID, err))
				return
			}
			if label == nil || label.Deleted {
				wpa.HandleNotFound(fmt.Sprintf("label %d not found", labelID))
				return
			}
		}
	}

	if err = controller.PolicyManager.UpdatePolicy(convertToWebhookPolicy(policy)); err != nil {
		wpa.HandleInternalServerError(fmt.Sprintf("failed to update policy %d: %v", id, err))
		return
	}
}

// Delete the webhook policy
func (wpa *WebhookPolicyAPI) Delete() {
	id := wpa.GetIDFromURL()

	policy, err := controller.PolicyManager.GetPolicy(id)
	if err != nil {
		log.Errorf("failed to get policy %d: %v", id, err)
		wpa.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if policy.ID == 0 {
		wpa.HandleNotFound(fmt.Sprintf("policy %d not found", id))
		return
	}

	if _, err := wpa.doAuth(policy.ProjectID); err != nil {
		return
	}

	count, err := dao.GetTotalCountOfWebhookJobs(&models.WebhookJobQuery{
		PolicyID: id,
		Statuses: []string{models.JobRunning, models.JobRetrying, models.JobPending},
	})
	if err != nil {
		log.Errorf("failed to filter jobs of policy %d: %v", id, err)
		wpa.CustomAbort(http.StatusInternalServerError, "")
	}
	if count > 0 {
		wpa.CustomAbort(http.StatusPreconditionFailed, "policy has running/retrying/pending jobs, can not be deleted")
	}

	// maybe this will make the jobs related to this policy to orphan
	// however removing policy is a logical operate, this makes stuffs acceptable
	if err = controller.PolicyManager.RemovePolicy(id); err != nil {
		log.Errorf("failed to delete policy %d: %v", id, err)
		wpa.CustomAbort(http.StatusInternalServerError, "")
	}
}

func (wpa *WebhookPolicyAPI) doAuth(projectID int64) (*models.Project, error) {
	project, err := wpa.ProjectMgr.Get(projectID)
	if err != nil {
		wpa.ParseAndHandleError(fmt.Sprintf("failed to get project %d", projectID), err)
		return nil, err
	}
	if project == nil {
		wpa.HandleNotFound(fmt.Sprintf("project %d not found", projectID))
		return nil, errors.New("project not found")
	}

	if !(wpa.Ctx.Input.IsGet() && wpa.SecurityCtx.HasReadPerm(projectID) ||
		wpa.SecurityCtx.HasAllPerm(projectID)) {
		wpa.HandleForbidden(wpa.SecurityCtx.GetUsername())
		return nil, errors.New("forbidden")
	}
	return project, nil
}

func convertToAPIModel(project *models.Project, policy *webhook_models.WebhookPolicy) (*api_models.WebhookPolicy, error) {
	if policy.ID == 0 {
		return nil, nil
	}

	// populate simple properties
	ply := &api_models.WebhookPolicy{
		ID:                policy.ID,
		Name:              policy.Name,
		Description:       policy.Description,
		Project:           project,
		HookTypes:         policy.HookTypes,
		Target:            policy.Target,
		CreationTime:      policy.CreationTime,
		UpdateTime:        policy.UpdateTime,
		Enabled:           policy.Enabled,
	}

	// populate label used in label filter
	for _, filter := range policy.Filters {
		if filter.Kind == replication.FilterItemKindLabel {
			labelID := filter.Value.(int64)
			label, err := dao.GetLabel(labelID)
			if err != nil {
				return nil, err
			}
			filter.Value = label
		}
		ply.Filters = append(ply.Filters, filter)
	}

	// get count of err job
	errJobCount, err := dao.GetTotalCountOfWebhookJobs(&models.WebhookJobQuery{
		PolicyID: policy.ID,
		Statuses: []string{models.JobError},
	})
	if err != nil {
		return nil, err
	}
	ply.ErrorJobCount = errJobCount

	return ply, nil
}

func convertToWebhookPolicy(policy *api_models.WebhookPolicy) webhook_models.WebhookPolicy {
	if policy == nil {
		return webhook_models.WebhookPolicy{}
	}

	ply := webhook_models.WebhookPolicy{
		ID:                policy.ID,
		Name:              policy.Name,
		Description:       policy.Description,
		Filters:           policy.Filters,
		HookTypes:         policy.HookTypes,
		Enabled:           policy.Enabled,
		Target:            policy.Target,
		ProjectID:         policy.Project.ProjectID,
		CreationTime:      policy.CreationTime,
		UpdateTime:        policy.UpdateTime,
	}
	return ply
}
