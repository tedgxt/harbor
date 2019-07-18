package event

import (
	"time"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/webhook/model"
)

// Event ...
type Event struct {
	HookType    string
	ProjectID   int64
	ProjectName string
	Payload     *model.Payload
}

// ImageEvent ...
type ImageEvent struct {
	Project  *models.Project
	Events   []*models.Event
	OccurAt  time.Time
	Operator string
	RepoName string
}
