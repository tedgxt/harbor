package job

import (
	"github.com/goharbor/harbor/src/webhook/models"
	filter_models "github.com/goharbor/harbor/src/replication/models"
)

// Interface is certain mechanism to generate JobData.
type Interface interface {
	Generate(policy *models.WebhookPolicy, triggerItems []filter_models.FilterItem) (*models.JobData, error)
}
