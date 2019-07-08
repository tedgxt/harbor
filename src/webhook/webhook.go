package webhook

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	cfg "github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/webhook/config"
	"github.com/goharbor/harbor/src/webhook/hook"
	"github.com/goharbor/harbor/src/webhook/model"
	"github.com/goharbor/harbor/src/webhook/operation"
	"github.com/goharbor/harbor/src/webhook/policy"
	"github.com/goharbor/harbor/src/webhook/policy/impl"
)

var (
	// PolicyManager is a global webhook policy manager
	PolicyManager policy.Manager
	// 	HookManager is a hook manager
	HookManager hook.Manager
	// ExecutionCtl is a webhook execution controller
	ExecutionCtl operation.Controller

	// SupportedHookTypes is a map store support webhook type, eg. pushImage, pullImage etc
	SupportedHookTypes map[string]int
)

func Init() {
	config.Config = &config.Configuration{
		CoreURL:          cfg.InternalCoreURL(),
		TokenServiceURL:  cfg.InternalTokenServiceEndpoint(),
		JobserviceURL:    cfg.InternalJobServiceURL(),
		CoreSecret:       cfg.CoreSecret(),
		JobserviceSecret: cfg.JobserviceSecret(),
	}

	// init webhook policy manager
	PolicyManager = impl.NewDefaultManger()
	// init hook manager
	HookManager = hook.NewHookManager()
	// init webhook execution controller
	ExecutionCtl = operation.NewController()

	SupportedHookTypes = make(map[string]int)

	initSupportedWebhookType(model.EventTypePushImage, model.EventTypePullImage, model.EventTypeDeleteImage,
		model.EventTypeUploadChart, model.EventTypeDeleteChart, model.EventTypeDownloadChart)

	log.Info("webhook initialization completed")
}

func initSupportedWebhookType(hookTypes ...string) {
	for _, hookType := range hookTypes {
		SupportedHookTypes[hookType] = model.ValidType
	}
}
