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
	"github.com/goharbor/harbor/src/replication/models"
)

// WebhookPolicy defines the structure of a webhook policy. This struct is used internally.
// Could transfer from dao model or transfer to api model.
type WebhookPolicy struct {
	ID                int64 // UUID of the policy
	Name              string
	Description       string
	Filters           []models.Filter
	ProjectID         int64  // Project attached to this policy
	Target            string
	HookTypes         []string
	CreationTime      time.Time
	UpdateTime        time.Time
	Enabled           bool
}
