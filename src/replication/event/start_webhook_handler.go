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

package event

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/goharbor/harbor/src/replication/event/notification"
	"github.com/goharbor/harbor/src/webhook/controller"
)

// StartWebhookHandler implements the notification handler interface to handle start webhook requests.
type StartWebhookHandler struct{}

// Handle implements the same method of notification handler interface
func (srh *StartWebhookHandler) Handle(value interface{}) error {
	if value == nil {
		return errors.New("StartWebhookHandler can not handle nil value")
	}

	vType := reflect.TypeOf(value)
	if vType.Kind() != reflect.Struct || vType.String() != "notification.StartWebhookNotification" {
		return fmt.Errorf("Mismatch value type of StartWebhookHandler, expect %s but got %s", "notification.StartWebhookNotification", vType.String())
	}

	notification := value.(notification.StartWebhookNotification)
	if notification.Policy == nil {
		return errors.New("Invalid policy")
	}

	// Start webhook
	return controller.HookManager.StartHook(notification.Policy , notification.Metadata)
}

// IsStateful implements the same method of notification handler interface
func (srh *StartWebhookHandler) IsStateful() bool {
	// Stateless
	return false
}

