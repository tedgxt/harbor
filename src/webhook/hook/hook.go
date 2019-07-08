package hook

import (
	"encoding/json"
	"fmt"
	"time"

	cJob "github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/job/models"
	cModels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/utils"
	"github.com/goharbor/harbor/src/webhook/config"
	"github.com/goharbor/harbor/src/webhook/execution"
	"github.com/goharbor/harbor/src/webhook/execution/impl"
	"github.com/goharbor/harbor/src/webhook/scheduler"
)

type Manager interface {
	StartHook(item *scheduler.ScheduleItem, data *models.JobData) error
}

type HookManager struct {
	execMgr execution.Manager
	client  cJob.Client
}

func NewHookManager() *HookManager {
	return &HookManager{
		execMgr: impl.NewDefaultManager(),
		client:  utils.GetJobServiceClient(),
	}
}

func (hm *HookManager) StartHook(item *scheduler.ScheduleItem, data *models.JobData) error {
	payload, err := json.Marshal(item.Payload)
	if err != nil {
		return err
	}

	t := time.Now()
	id, err := hm.execMgr.Create(&cModels.WebhookExecution{
		PolicyID:     item.PolicyId,
		HookType:     item.Target.Type,
		Status:       cModels.ExecutionStatusInProgress,
		CreationTime: t,
		UpdateTime:   t,
		JobDetail:    string(payload),
	})
	if err != nil {
		return fmt.Errorf("failed to create the execution record for webhook based on policy %d: %v", item.PolicyId, err)
	}
	statusHookURL := fmt.Sprintf("%s/service/notifications/jobs/webhook/%d", config.Config.CoreURL, id)
	data.StatusHook = statusHookURL

	log.Debugf("created a webhook execution %d for the policy %d", id, item.PolicyId)

	// submit hook job to jobservice
	go func() {
		whExecution := &cModels.WebhookExecution{
			ID:         id,
			UpdateTime: time.Now(),
		}

		jobUUID, err := hm.client.SubmitJob(data)
		if err != nil {
			log.Errorf("failed to process the webhook event: %v", err)
			e := hm.execMgr.Update(whExecution, "Status", "UpdateTime")
			if e != nil {
				log.Errorf("failed to update the webhook execution %d: %v", id, e)
			}
			return
		}
		whExecution.UUID = jobUUID
		e := hm.execMgr.Update(whExecution, "JobUuid", "UpdateTime")
		if e != nil {
			log.Errorf("failed to update the webhook execution %d: %v", id, e)
		}
	}()
	return nil
}
