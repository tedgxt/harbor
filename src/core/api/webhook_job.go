package api

import (
	"errors"
	"fmt"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/webhook"
)

// WebhookJobAPI ...
type WebhookJobAPI struct {
	BaseController
	project *models.Project
}

// Prepare ...
func (w *WebhookJobAPI) Prepare() {
	w.BaseController.Prepare()
	if !w.SecurityCtx.IsAuthenticated() {
		w.SendUnAuthorizedError(errors.New("UnAuthorized"))
		return
	}

	pid, err := w.GetInt64FromPath(":pid")
	if err != nil {
		w.SendBadRequestError(fmt.Errorf("failed to get project ID: %v", err))
		return
	}
	if pid <= 0 {
		w.SendBadRequestError(fmt.Errorf("invalid project ID: %d", pid))
		return
	}

	project, err := w.ProjectMgr.Get(pid)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to get project %d: %v", pid, err))
		return
	}
	if project == nil {
		w.SendNotFoundError(fmt.Errorf("project %d not found", pid))
		return
	}
	w.project = project
}

// List ...
func (w *WebhookJobAPI) List() {
	if !w.validateRBAC(rbac.ActionList, w.project.ProjectID) {
		return
	}

	policyID, err := w.GetInt64("policy_id")
	if err != nil || policyID <= 0 {
		w.SendBadRequestError(fmt.Errorf("invalid policy_id: %s", w.GetString("policy_id")))
		return
	}

	policy, err := webhook.PolicyCtl.Get(policyID)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to get policy %d: %v", policyID, err))
		return
	}
	if policy == nil {
		w.SendBadRequestError(fmt.Errorf("policy %d not found", policyID))
		return
	}

	query := &models.WebhookJobQuery{
		PolicyID: policyID,
	}

	query.Statuses = w.GetStrings("status")

	query.Page, query.Size, err = w.GetPaginationParams()
	if err != nil {
		w.SendBadRequestError(err)
		return
	}

	total, jobs, err := webhook.JobCtl.ListWebhookJobs(query)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to list webhook jobs: %v", err))
		return
	}
	w.SetPaginationHeader(total, query.Page, query.Size)
	w.WriteJSONData(jobs)
}

func (w *WebhookJobAPI) validateRBAC(action rbac.Action, projectID int64) bool {
	if w.SecurityCtx.IsSysAdmin() {
		return true
	}

	project, err := w.ProjectMgr.Get(projectID)
	if err != nil {
		w.ParseAndHandleError(fmt.Sprintf("failed to get project %d", projectID), err)
		return false
	}

	resource := rbac.NewProjectNamespace(project.ProjectID).Resource(rbac.ResourceWebhookPolicy)
	if !w.SecurityCtx.Can(action, resource) {
		w.SendForbiddenError(errors.New(w.SecurityCtx.GetUsername()))
		return false
	}
	return true
}
