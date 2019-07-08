package scheduler

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/notifier"
	"github.com/goharbor/harbor/src/webhook/model"
	"github.com/goharbor/harbor/src/webhook/scheduler/topic"
)

func init() {
	handlersMap := map[string][]notifier.NotificationHandler{
		topic.WebhookOperationTopicOnHTTP: {&HttpScheduler{}},
	}

	for t, handlers := range handlersMap {
		for _, handler := range handlers {
			if err := notifier.Subscribe(t, handler); err != nil {
				log.Errorf("failed to subscribe topic %s: %v", t, err)
				continue
			}
			log.Debugf("topic %s is subscribed", t)
		}
	}
}

// ScheduleItem is an item that can be scheduled
type ScheduleItem struct {
	PolicyId int64
	Target   *model.HookTarget
	Payload  interface{}
	IsTest   bool
}
