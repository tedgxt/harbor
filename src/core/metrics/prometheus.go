package metrics

import (
	"github.com/uber-go/tally"
	promreporter "github.com/uber-go/tally/prometheus"
	"io"
	"time"
)

func NewPrometheusScope(listenAddr string) (tally.Scope, io.Closer, error) {
	prometheusConfig := promreporter.Configuration{
		ListenAddress: listenAddr,
	}
	r, err := prometheusConfig.NewReporter(promreporter.ConfigurationOptions{})
	if err != nil {
		return nil, nil, err
	}
	s, c := tally.NewRootScope(tally.ScopeOptions{
		CachedReporter: r,
	}, time.Second)
	return s, c, nil
}
