package notification

import (
	"errors"
	"fmt"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/notifier/model"
	notifyModel "github.com/goharbor/harbor/src/core/notifier/model"
	"github.com/goharbor/harbor/src/pkg/notification"
)

// RetentionPreprocessHandler preprocess retention event data
type RetentionPreprocessHandler struct {
}

func (rp *RetentionPreprocessHandler) Handle(value interface{}) error {
	if !config.NotificationEnable() {
		log.Debug("notification feature is not enabled")
		return nil
	}
	retentionEvent, ok := value.(*model.RetentionEvent)
	if !ok {
		return errors.New("invalid quota event type")
	}
	if retentionEvent == nil {
		return fmt.Errorf("nil retention event")
	}

	project, err := config.GlobalProjectMgr.Get(retentionEvent.Project.Name)
	if err != nil {
		log.Errorf("failed to get project:%s, error: %v", retentionEvent.Project.Name, err)
		return err
	}
	if project == nil {
		return fmt.Errorf("project not found of retention event: %s", retentionEvent.Project.Name)
	}
	policies, err := notification.PolicyMgr.GetRelatedPolices(project.ProjectID, retentionEvent.EventType)
	if err != nil {
		log.Errorf("failed to find policy for %s event: %v", retentionEvent.EventType, err)
		return err
	}
	if len(policies) == 0 {
		log.Debugf("cannot find policy for %s event: %v", retentionEvent.EventType, retentionEvent)
		return nil
	}

	payload, err := constructRetentionPayload(retentionEvent)
	if err != nil {
		return err
	}

	err = sendHookWithPolicies(policies, payload, retentionEvent.EventType)
	if err != nil {
		return err
	}
	return nil
}

func constructRetentionPayload(event *model.RetentionEvent) (*model.Payload, error) {
	repository := event.Repository
	if repository == "" {
		return nil, fmt.Errorf("invalid %s event with empty repository", event.EventType)
	}

	repoType := models.ProjectPrivate
	if event.Project.IsPublic() {
		repoType = models.ProjectPublic
	}

	// TODO: current payload cannot carry all retention info
	payload := &notifyModel.Payload{
		Type:    event.EventType,
		OccurAt: event.OccurAt.Unix(),
		EventData: &notifyModel.EventData{
			Repository: &notifyModel.Repository{
				Name:         repository,
				Namespace:    event.Project.Name,
				RepoFullName: event.Project.Name + "/" + repository,
				RepoType:     repoType,
			},
		},
	}
	resource := &notifyModel.Resource{
		RetentionOverView: &notifyModel.RetentionOverView{
			Total:    event.ImageCount,
			Retained: event.Retained,
		},
	}
	payload.EventData.Resources = append(payload.EventData.Resources, resource)

	return payload, nil
}
