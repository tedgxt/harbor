package scheduler

import (
	"errors"

	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/webhook"
	"github.com/goharbor/harbor/src/webhook/hook"
)

type HttpScheduler struct {
}

func (h *HttpScheduler) Handle(value interface{}) error {
	if value == nil {
		return errors.New("HttpScheduler cannot handle nil value")
	}

	item, ok := value.(hook.ScheduleItem)
	if !ok {
		return errors.New("invalid webhook http schedule item")
	}

	return h.process(&item)
}

// IsStateful ...
func (h *HttpScheduler) IsStateful() bool {
	return false
}

func (h *HttpScheduler) process(item *hook.ScheduleItem) error {
	j := &models.JobData{
		Metadata: &models.JobMetadata{
			JobKind: job.KindGeneric,
		},
	}
	j.Name = job.WebhookHTTPJob

	j.Parameters = map[string]interface{}{
		"payload": item.Payload,
		"address": item.Target.Address,
		"secret":  item.Target.Secret,
	}
	return webhook.HookManager.StartHook(item, j)
}
