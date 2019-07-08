package operation

import (
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/webhook/execution"
	"github.com/goharbor/harbor/src/webhook/execution/impl"
)

// Controller handles the webhook related operations
type Controller interface {
	// GetWebhookExecution get a webhook execution
	GetWebhookExecution(int64) (*models.WebhookExecution, error)

	// ListWebhookExecutions list webhook executions
	ListWebhookExecutions(...*models.WebhookExecutionQuery) (int64, []*models.WebhookExecution, error)

	// UpdateWebhookExecution update webhook execution
	UpdateWebhookExecution(*models.WebhookExecution, ...string) error

	// DeleteWebhookExecution delete webhook execution
	DeleteWebhookExecution(int64) error
}

type controller struct {
	execMgr execution.Manager
}

// NewController returns a controller implementation
func NewController() Controller {
	ctl := &controller{
		execMgr: impl.NewDefaultManager(),
	}
	return ctl
}

// GetWebhookExecution ...
func (c *controller) GetWebhookExecution(id int64) (*models.WebhookExecution, error) {
	return c.execMgr.Get(id)
}

// ListWebhookExecutions ...
func (c *controller) ListWebhookExecutions(query ...*models.WebhookExecutionQuery) (int64, []*models.WebhookExecution, error) {
	return c.execMgr.List(query...)
}

// UpdateWebhookExecution ...
func (c *controller) UpdateWebhookExecution(execution *models.WebhookExecution, props ...string) error {
	return c.execMgr.Update(execution, props...)
}

// DeleteWebhookExecution ...
func (c *controller) DeleteWebhookExecution(id int64) error {
	return c.execMgr.Delete(id)
}

//func createWebhookExecution(mgr execution.Manager, policyID int64, event *event.Event) (int64, error) {
//	t := time.Now()
//	jobDetail, err := json.Marshal(event.Payload)
//	if err != nil {
//		return 0, err
//	}
//
//	id, err := mgr.Create(
//		&models.WebhookExecution{
//			PolicyID:     policyID,
//			HookType:     event.HookType,
//			Status:       models.ExecutionStatusInProgress,
//			CreationTime: t,
//			UpdateTime:   t,
//			JobDetail:    string(jobDetail),
//		})
//	if err != nil {
//		return 0, fmt.Errorf("failed to create the execution record for webhook based on policy %d: %v", policyID, err)
//	}
//	log.Debugf("an execution record for webhook based on the policy %d created: %d", policyID, id)
//	return id, nil
//}
