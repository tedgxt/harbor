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

package models

import (
	"time"
)

const (
	// WebhookJobTable is the table name for webhook jobs
	WebhookJobTable = "webhook_job"
	// WebhookPolicyTable is table name for webhook policies
	WebhookPolicyTable = "webhook_policy"
)

// WebhookPolicy is the model for a webhook policy, which associates a project with a target by filters and types
type WebhookPolicy struct {
	ID                int64     `orm:"pk;auto;column(id)"`
	ProjectID         int64     `orm:"column(project_id)" `
	Target            string    `orm:"column(target)"`
	HookTypes         string    `orm:"column(hook_types)"`
	Name              string    `orm:"column(name)"`
	Description       string    `orm:"column(description)"`
	Filters           string    `orm:"column(filters)"`
	CreationTime      time.Time `orm:"column(creation_time);auto_now_add"`
	UpdateTime        time.Time `orm:"column(update_time);auto_now"`
	Enabled           bool      `orm:"column(enabled)"`
	Deleted           bool      `orm:"column(deleted)"`
}

// WebhookJob is the model for a webhook job, which is the execution unit on job service,
// currently it is used to trigger a hook to a remote endpoint by a http request
type WebhookJob struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	Status       string    `orm:"column(status)" json:"status"`
	PolicyID     int64     `orm:"column(policy_id)" json:"policy_id"`
	HookType     string    `orm:"column(hook_type)" json:"hook_type"`
	JobDetail    string    `orm:"column(job_detail)" json:"job_detail"`
	UUID         string    `orm:"column(job_uuid)" json:"-"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// TableName is required by by beego orm to map WebhookPolicy to table webhook_policy
func (r *WebhookPolicy) TableName() string {
	return WebhookPolicyTable
}

// TableName is required by by beego orm to map WebhookJob to table webhook_job
func (r *WebhookJob) TableName() string {
	return WebhookJobTable
}

// WebhookJobQuery holds query conditions for webhook job
type WebhookJobQuery struct {
	PolicyID   int64
	Statuses   []string
	HookTypes  []string
	StartTime  *time.Time
	EndTime    *time.Time
	Pagination
}
