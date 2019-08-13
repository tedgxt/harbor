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
	"time"

	"github.com/goharbor/harbor/src/common/dao"
	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/utils"
	"github.com/goharbor/harbor/src/webhook/controller"
	"errors"
)

// RepJobAPI handles request to /api/jobs/webhook /api/jobs/webhook/:id/log
type WebhookJobAPI struct {
	BaseController
	jobID int64
}

// Prepare validates that whether user has system admin role
func (wja *WebhookJobAPI) Prepare() {
	wja.BaseController.Prepare()
	if !wja.SecurityCtx.IsAuthenticated() {
		wja.HandleUnauthorized()
		return
	}

	if len(wja.GetStringFromPath(":id")) != 0 {
		id, err := wja.GetInt64FromPath(":id")
		if err != nil {
			wja.HandleBadRequest(fmt.Sprintf("invalid ID: %s", wja.GetStringFromPath(":id")))
			return
		}
		wja.jobID = id
	}
}

// List filters jobs according to the parameters
func (wja *WebhookJobAPI) List() {
	policyID, err := wja.GetInt64("policy_id")
	if err != nil || policyID <= 0 {
		wja.HandleBadRequest(fmt.Sprintf("invalid policy_id: %s", wja.GetString("policy_id")))
		return
	}

	policy, err := controller.PolicyManager.GetPolicy(policyID)
	if err != nil {
		log.Errorf("failed to get policy %d: %v", policyID, err)
		wja.CustomAbort(http.StatusInternalServerError, "")
	}

	if policy.ID == 0 {
		wja.HandleNotFound(fmt.Sprintf("policy %d not found", policyID))
		return
	}

	if _, err := wja.doAuth(policy.ProjectID); err != nil {
		return
	}

	query := &models.WebhookJobQuery{
		PolicyID: policyID,
	}

	query.Statuses = wja.GetStrings("status")
	//query.OpUUID = wja.GetString("op_uuid")

	startTimeStr := wja.GetString("start_time")
	if len(startTimeStr) != 0 {
		i, err := strconv.ParseInt(startTimeStr, 10, 64)
		if err != nil {
			wja.HandleBadRequest(fmt.Sprintf("invalid start_time: %s", startTimeStr))
			return
		}
		t := time.Unix(i, 0)
		query.StartTime = &t
	}

	endTimeStr := wja.GetString("end_time")
	if len(endTimeStr) != 0 {
		i, err := strconv.ParseInt(endTimeStr, 10, 64)
		if err != nil {
			wja.HandleBadRequest(fmt.Sprintf("invalid end_time: %s", endTimeStr))
			return
		}
		t := time.Unix(i, 0)
		query.EndTime = &t
	}

	query.Page, query.Size = wja.GetPaginationParams()

	total, err := dao.GetTotalCountOfWebhookJobs(query)
	if err != nil {
		wja.HandleInternalServerError(fmt.Sprintf("failed to get total count of webhook jobs of policy %d: %v", policyID, err))
		return
	}
	jobs, err := dao.GetWebhookJobs(query)
	if err != nil {
		wja.HandleInternalServerError(fmt.Sprintf("failed to get webhook jobs, query: %v :%v", query, err))
		return
	}

	wja.SetPaginationHeader(total, query.Page, query.Size)

	wja.Data["json"] = jobs
	wja.ServeJSON()
}

// Delete ...
func (wja *WebhookJobAPI) Delete() {
	if wja.jobID == 0 {
		wja.HandleBadRequest("ID is nil")
		return
	}

	job, err := dao.GetWebhookJob(wja.jobID)
	if err != nil {
		log.Errorf("failed to get job %d: %v", wja.jobID, err)
		wja.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	// check job
	if job == nil {
		wja.HandleNotFound(fmt.Sprintf("job %d not found", wja.jobID))
		return
	}
	if job.Status == models.JobPending || job.Status == models.JobRunning {
		wja.HandleBadRequest(fmt.Sprintf("job is %s, can not be deleted", job.Status))
		return
	}

	// admin can delete all jobs including jobs without policy
	if !wja.SecurityCtx.IsSysAdmin() {
		policy, err := controller.PolicyManager.GetPolicy(job.PolicyID)
		if err != nil {
			wja.ParseAndHandleError(fmt.Sprintf("failed to get webhook policy %d", job.PolicyID), err)
			return
		}
		if policy.ID == 0 {
			wja.HandleNotFound(fmt.Sprintf("policy %d not found", job.PolicyID))
			return
		}
		if _, err := wja.doAuth(policy.ProjectID); err != nil {
			return
		}
	}

	if err = dao.DeleteWebhookJob(wja.jobID); err != nil {
		log.Errorf("failed to deleted job %d: %v", wja.jobID, err)
		wja.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

// GetLog ...
func (wja *WebhookJobAPI) GetLog() {
	if wja.jobID == 0 {
		wja.HandleBadRequest("ID is nil")
		return
	}

	job, err := dao.GetWebhookJob(wja.jobID)
	if err != nil {
		wja.HandleInternalServerError(fmt.Sprintf("failed to get webhook job %d: %v", wja.jobID, err))
		return
	}

	if job == nil {
		wja.HandleNotFound(fmt.Sprintf("webhook job %d not found", wja.jobID))
		return
	}

	// admin can get all job logs including jobs without policy
	if !wja.SecurityCtx.IsSysAdmin() {
		policy, err := controller.PolicyManager.GetPolicy(job.PolicyID)
		if err != nil {
			wja.ParseAndHandleError(fmt.Sprintf("failed to get webhook policy %d", job.PolicyID), err)
			return
		}
		if policy.ID == 0 {
			wja.HandleNotFound(fmt.Sprintf("policy %d not found", job.PolicyID))
			return
		}
		if _, err := wja.doAuth(policy.ProjectID); err != nil {
			return
		}
	}

	logBytes, err := utils.GetJobServiceClient().GetJobLog(job.UUID)
	if err != nil {
		if httpErr, ok := err.(*common_http.Error); ok {
			wja.RenderError(httpErr.Code, "")
			log.Errorf(fmt.Sprintf("failed to get log of job %d: %d %s",
				wja.jobID, httpErr.Code, httpErr.Message))
			return
		}
		wja.HandleInternalServerError(fmt.Sprintf("failed to get log of job %s: %v",
			job.UUID, err))
		return
	}
	wja.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Length"), strconv.Itoa(len(logBytes)))
	wja.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Type"), "text/plain")
	_, err = wja.Ctx.ResponseWriter.Write(logBytes)
	if err != nil {
		wja.HandleInternalServerError(fmt.Sprintf("failed to write log of job %s: %v", job.UUID, err))
		return
	}
}

func (wja *WebhookJobAPI) doAuth(projectID int64) (*models.Project, error) {
	project, err := wja.ProjectMgr.Get(projectID)
	if err != nil {
		wja.ParseAndHandleError(fmt.Sprintf("failed to get project %d", projectID), err)
		return nil, err
	}
	if project == nil {
		wja.HandleNotFound(fmt.Sprintf("project %d not found", projectID))
		return nil, errors.New("project not found")
	}

	if !(wja.Ctx.Input.IsGet() && wja.SecurityCtx.HasReadPerm(projectID) ||
		wja.SecurityCtx.HasAllPerm(projectID)) {
		wja.HandleForbidden(wja.SecurityCtx.GetUsername())
		return nil, errors.New("forbidden")
	}
	return project, nil
}
