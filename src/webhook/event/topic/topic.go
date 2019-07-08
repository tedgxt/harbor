package topic

const (
	// WebhookEventTopicOnImage include pushImage, pullImage, deleteImage
	WebhookEventTopicOnImage = "OnImage"

	// WebhookEventTopicOnChart include uploadChart, deleteChart, downloadChart
	WebhookEventTopicOnChart = "OnChart"

	// WebhookEventTopicOnScan include scanningFailed, scanningCompleted
	WebhookEventTopicOnScan = "OnScan"
)
