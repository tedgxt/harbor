package job

import (
	"github.com/goharbor/harbor/src/webhook/models"
	filter_models "github.com/goharbor/harbor/src/replication/models"
	"fmt"
	"github.com/goharbor/harbor/src/core/config"
	"time"
	"github.com/goharbor/harbor/src/webhook"
	"strings"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils"
	coreutils "github.com/goharbor/harbor/src/core/utils"
	"github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/docker/distribution/manifest/schema2"
)

const (
	Project_Private  = "private"
	Project_Public   = "public"
)

// PushJobGenerator ...
type PushJobGenerator struct {
}

func NewPushJobGenerator() *PushJobGenerator {
	return &PushJobGenerator{}
}

func (pjg *PushJobGenerator) Generate(policy *models.WebhookPolicy, triggerItems []filter_models.FilterItem) (*models.JobData, error) {
	project, err := config.GlobalProjectMgr.Get(policy.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %d : %v", policy.ProjectID, err)
	}
	var events = []interface{}{}
	for _, item := range triggerItems {
		strs := strings.SplitN(item.Value, ":", 2)
		if len(strs) != 2 {
			log.Warningf("invalid image '%s'", item.Value)
			continue
		}
		_, repoName := utils.ParseRepository(strs[0])
		projectType := Project_Private
		if project.IsPublic() {
			projectType = Project_Public
		}
		extURL, _ := config.ExtURL()
		event := models.PushEvent{
			Project:     project.Name,
			RepoName:    repoName,
			Tag:         strs[1],
			FullName:    strs[0],
			TriggerTime: time.Now().UTC(),
			Digest:      getDigest(strs[0], strs[1]),
			ProjectType: projectType,
			ResourceURL: fmt.Sprintf("%s/%s", extURL, item.Value),
		}
		events = append(events, event)
	}
	return &models.JobData{
		EventType: webhook.ImagePushEvent,
		Events:    events,
	}, nil
}

func getImageID(repoName, tag string) string {
	client, err := coreutils.NewRepositoryClientForUI("harbor-core", repoName)
	if err != nil {
		log.Errorf("failed to create repository client: %v", err)
		return ""
	}
	mediaTypes := []string{}
	mediaTypes = append(mediaTypes, schema2.MediaTypeManifest)
	_, mediaType, payload, err := client.PullManifest(tag, mediaTypes)
	if err != nil {
		log.Errorf("failed to pull manifest: %v", err)
		return ""
	}
	manifest, _, err := registry.UnMarshal(mediaType, payload)
	if err != nil {
		log.Errorf("failed to unmarshal manifest payload: %v", err)
		return ""
	}

	deserializedmanifest, ok := manifest.(*schema2.DeserializedManifest)
	if ok {
		return deserializedmanifest.Target().Digest.String()
	}
	return ""
}

func getDigest(repoName, tag string) string {
	client, err := coreutils.NewRepositoryClientForUI("harbor-core", repoName)
	if err != nil {
		log.Errorf("failed to create repository client: %v", err)
		return ""
	}
	mediaTypes := []string{}
	mediaTypes = append(mediaTypes, schema2.MediaTypeManifest)
	digest, _, _, err := client.PullManifest(tag, mediaTypes)
	if err != nil {
		log.Errorf("failed to pull manifest: %v", err)
		return ""
	}
	return digest
}
