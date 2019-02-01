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

package models

import (
	"time"

	"github.com/astaxie/beego/validation"
	common_models "github.com/goharbor/harbor/src/common/models"
	rep_models "github.com/goharbor/harbor/src/replication/models"
)

// WebhookPolicy defines the data model used in API level
type WebhookPolicy struct {
	ID                        int64                      `json:"id"`
	Name                      string                     `json:"name"`
	Description               string                     `json:"description"`
	Filters                   []rep_models.Filter        `json:"filters"`
	Project                   *common_models.Project     `json:"project"`
	Target                    string                     `json:"target"`
	HookTypes                 []string                   `json:"hook_types"`
	CreationTime              time.Time                  `json:"creation_time"`
	UpdateTime                time.Time                  `json:"update_time"`
	Enabled                   bool                       `json:"enabled"`
	ErrorJobCount             int64                      `json:"error_job_count"`
}

// Valid ...
func (w *WebhookPolicy) Valid(v *validation.Validation) {
	if len(w.Name) == 0 {
		v.SetError("name", "can not be empty")
	}

	if len(w.Name) > 256 {
		v.SetError("name", "max length is 256")
	}

	if len(w.HookTypes) == 0 {
		v.SetError("hook types", "can not be empty")
	}

	if len(w.Target) == 0 {
		v.SetError("target", "can not be empty")
	}

	if len(w.Target) > 512 {
		v.SetError("target", "max length is 512")
	}

	for i := range w.Filters {
		w.Filters[i].Valid(v)
	}
}

