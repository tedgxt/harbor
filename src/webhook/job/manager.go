package job

import (
   "github.com/goharbor/harbor/src/webhook/models"
   filter_models "github.com/goharbor/harbor/src/replication/models"
   "github.com/goharbor/harbor/src/webhook"
   "fmt"
)

// Manager defines the method a job manger should implement
type Manager interface {
   Generate(policy *models.WebhookPolicy, triggerItems []filter_models.FilterItem) (*models.JobData, error)
}

// DefaultManager ...
type DefaultManager struct{
   handlers map[string]Interface
}

// NewDefaultManager is the constructor of DefaultManager.
func NewDefaultManager() *DefaultManager {
   return &DefaultManager{
      handlers: map[string]Interface{
         webhook.ImagePushEvent: NewPushJobGenerator(),
      },
   }
}

// Generate returns the webhook request body data(job data)
func (dm *DefaultManager) Generate(policy *models.WebhookPolicy, triggerItems []filter_models.FilterItem) (*models.JobData, error) {
   if policy == nil {
      return nil, fmt.Errorf("generate job data: policy is nil")
   }
   if len(triggerItems) == 0 {
      return nil, fmt.Errorf("generate job data: trigger items are nil")
   }

   hookType := triggerItems[0].Operation
   handler := dm.handlers[hookType]
   if handler == nil {
      return nil, fmt.Errorf("generate job data: handler doesn't exist")
   }

   return handler.Generate(policy, triggerItems)
}
