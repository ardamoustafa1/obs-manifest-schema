package manifest

import "fmt"

const (
	APIVersionV1 = "obs.acme.com/v1"
	APIVersionV2 = "obs.acme.com/v2"
	Kind         = "ServiceObservability"
)

type ServiceObservability struct {
	APIVersion string   `json:"apiVersion" yaml:"apiVersion"`
	Kind       string   `json:"kind" yaml:"kind"`
	Metadata   Metadata `json:"metadata" yaml:"metadata"`
	Spec       Spec     `json:"spec" yaml:"spec"`
}

type Metadata struct {
	Name             string `json:"name" yaml:"name"`
	Owner            string `json:"owner,omitempty" yaml:"owner,omitempty"`
	SlackChannel     string `json:"slack_channel,omitempty" yaml:"slack_channel,omitempty"`
	PagerDutyService string `json:"pagerduty_service,omitempty" yaml:"pagerduty_service,omitempty"`
}

type Spec struct {
	ServiceType   string        `json:"serviceType" yaml:"serviceType"`
	Environment   string        `json:"environment" yaml:"environment"`
	SLO           SLO           `json:"slo" yaml:"slo"`
	GoldenSignals GoldenSignals `json:"goldenSignals" yaml:"goldenSignals"`
	Dashboards    Dashboards    `json:"dashboards" yaml:"dashboards"`
	Alerting      Alerting      `json:"alerting" yaml:"alerting"`
	Ownership     *Ownership    `json:"ownership,omitempty" yaml:"ownership,omitempty"`
}

type Ownership struct {
	Team             string `json:"team" yaml:"team"`
	SlackChannel     string `json:"slackChannel" yaml:"slackChannel"`
	PagerDutyService string `json:"pagerDutyService" yaml:"pagerDutyService"`
}

type SLO struct {
	AvailabilityTarget float64 `json:"availability_target" yaml:"availability_target"`
	LatencyTargetMS    int     `json:"latency_target_ms" yaml:"latency_target_ms"`
	LatencyPercentile  int     `json:"latency_percentile" yaml:"latency_percentile"`
}

type GoldenSignals struct {
	TrafficMetric            string   `json:"trafficMetric" yaml:"trafficMetric"`
	ErrorMetric              string   `json:"errorMetric" yaml:"errorMetric"`
	LatencyMetric            string   `json:"latencyMetric,omitempty" yaml:"latencyMetric,omitempty"`
	ProcessingDurationMetric string   `json:"processingDurationMetric,omitempty" yaml:"processingDurationMetric,omitempty"`
	SaturationMetrics        []string `json:"saturationMetrics" yaml:"saturationMetrics"`
}

type Dashboards struct {
	Template string `json:"template" yaml:"template"`
}

type Alerting struct {
	BurnRateAlerts bool        `json:"burnRateAlerts" yaml:"burnRateAlerts"`
	ExtraRules     []AlertRule `json:"extraRules" yaml:"extraRules"`
}

type AlertRule struct {
	Name     string `json:"name" yaml:"name"`
	Expr     string `json:"expr" yaml:"expr"`
	For      string `json:"for" yaml:"for"`
	Severity string `json:"severity" yaml:"severity"`
	Summary  string `json:"summary" yaml:"summary"`
}

type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

type Issue struct {
	Severity Severity
	Path     string
	Message  string
}

func (i Issue) String() string {
	if i.Path == "" {
		return fmt.Sprintf("%s: %s", i.Severity, i.Message)
	}
	return fmt.Sprintf("%s: %s: %s", i.Severity, i.Path, i.Message)
}

type Report struct {
	APIVersion string
	Issues     []Issue
}

func (r Report) Valid() bool {
	for _, issue := range r.Issues {
		if issue.Severity == SeverityError {
			return false
		}
	}
	return true
}

func (r Report) HasWarnings() bool {
	for _, issue := range r.Issues {
		if issue.Severity == SeverityWarning {
			return true
		}
	}
	return false
}
