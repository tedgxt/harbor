package webhook

import (
	"bytes"
	"crypto/tls"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"net/http"
)

const (
	secretHeaderName        = "Authorization"
	secretHeaderValuePrefix = "Secret"
	defaultMaxFails         = 5
)

type HttpNotifier struct {
	client *http.Client
	logger logger.Interface
	ctx    job.Context
}

// MaxFails returns that how many times this job can fail, get this value from config manager.
func (hn *HttpNotifier) MaxFails() uint {
	if maxFails, ok := hn.ctx.Get("webhook_max_retry"); ok {
		if maxFailsV, yes := maxFails.(uint); yes {
			return maxFailsV
		}

		return defaultMaxFails
	}

	return defaultMaxFails
}

// ShouldRetry ...
func (hn *HttpNotifier) ShouldRetry() bool {
	return true
}

// Validate implements the interface in job/Interface
func (hn *HttpNotifier) Validate(params job.Parameters) error {
	return nil
}

// Run implements the interface in job/Interface
func (hn *HttpNotifier) Run(ctx job.Context, params job.Parameters) error {
	if err := hn.init(ctx, params); err != nil {
		return err
	}

	err := hn.execute(ctx, params)
	return err
}

// init http_notifier for webhoook
func (hn *HttpNotifier) init(ctx job.Context, params map[string]interface{}) error {
	hn.logger = ctx.GetLogger()
	hn.ctx = ctx
	hn.client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			// when sending notification by https, skip verifying certificate
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		},
	}

	return nil
}

// send notification by http or https
func (hn *HttpNotifier) execute(ctx job.Context, params map[string]interface{}) error {
	payload := params["payload"].(string)
	address := params["address"].(string)

	req, err := http.NewRequest(http.MethodPost, address, bytes.NewReader([]byte(payload)))
	if params["secret"] != nil {
		secret := params["secret"].(string)
		req.Header.Set(secretHeaderName, secretHeaderValuePrefix+secret)
	}

	if err != nil {
		return err
	}
	resp, err := hn.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 && resp.StatusCode >= 300 {
		hn.logger.Errorf("webhook job(target: %s, jobData: %s) response code is %d", address, payload, resp.StatusCode)
	}

	return nil
}
