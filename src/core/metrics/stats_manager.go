package metrics

import "github.com/uber-go/tally"

const (
	PrometheusStats = "prometheus"
)

var GlobalStatsManager = &StatsManager{
	stats: make(map[string]tally.Scope),
}

type StatsManager struct {
	stats map[string]tally.Scope
}

func (sm *StatsManager) Register(statsType string, stat tally.Scope) {
	sm.stats[statsType] = stat
}

func (sm *StatsManager) GetStat(statsType string) tally.Scope {
	return sm.stats[statsType]
}
