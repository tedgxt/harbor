package webhook

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"net/http"
)

type HttpNotifier struct {
	client *http.Client
	logger logger.Interface
	ctx    job.Context
}

// MaxFails returns that how many times this job can fail, get this value from ctx.
func (hn *HttpNotifier) MaxFails() uint {
	// Max retry interval is around 3h
	// Large enough to ensure most situations can notify successfully
	return 10
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
	if _, ok := params["secret"]; ok {
		secret := params["secret"].(string)
		req.Header.Set("Authorization", "Secret" + secret)
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
		return fmt.Errorf("webhook job(target: %s) response code is %d", address, resp.StatusCode)
	}

	return nil
}
