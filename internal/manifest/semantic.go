package manifest

import (
	"fmt"
	"regexp"
	"strings"
)

var metricNamePattern = regexp.MustCompile(`^[a-z_:][a-z0-9_:]*$`)

func semanticIssues(manifest *ServiceObservability) []Issue {
	issues := make([]Issue, 0)
	percentile := manifest.Spec.SLO.LatencyPercentile
	if percentile != 50 && percentile != 90 && percentile != 95 && percentile != 99 {
		issues = append(issues, Issue{
			Severity: SeverityError,
			Path:     "/spec/slo/latency_percentile",
			Message:  "must be one of 50, 90, 95, or 99",
		})
	}

	workerLike := manifest.Spec.ServiceType == "worker" || manifest.Spec.ServiceType == "batch"
	if workerLike {
		if manifest.Spec.GoldenSignals.ProcessingDurationMetric == "" {
			issues = append(issues, Issue{Severity: SeverityError, Path: "/spec/goldenSignals/processingDurationMetric", Message: "is required for worker and batch services"})
		}
		if manifest.Spec.GoldenSignals.LatencyMetric != "" {
			issues = append(issues, Issue{Severity: SeverityError, Path: "/spec/goldenSignals/latencyMetric", Message: "is not allowed for worker and batch services; use processingDurationMetric"})
		}
	} else {
		if manifest.Spec.GoldenSignals.LatencyMetric == "" {
			issues = append(issues, Issue{Severity: SeverityError, Path: "/spec/goldenSignals/latencyMetric", Message: "is required for http, grpc, and database services"})
		}
		if manifest.Spec.GoldenSignals.ProcessingDurationMetric != "" {
			issues = append(issues, Issue{Severity: SeverityError, Path: "/spec/goldenSignals/processingDurationMetric", Message: "is only allowed for worker and batch services"})
		}
	}

	issues = append(issues, metricConventionIssues("/spec/goldenSignals/trafficMetric", manifest.Spec.GoldenSignals.TrafficMetric, []string{"_total"})...)
	issues = append(issues, metricConventionIssues("/spec/goldenSignals/errorMetric", manifest.Spec.GoldenSignals.ErrorMetric, []string{"_total"})...)
	if manifest.Spec.GoldenSignals.LatencyMetric != "" {
		issues = append(issues, metricConventionIssues("/spec/goldenSignals/latencyMetric", manifest.Spec.GoldenSignals.LatencyMetric, []string{"_seconds"})...)
	}
	if manifest.Spec.GoldenSignals.ProcessingDurationMetric != "" {
		issues = append(issues, metricConventionIssues("/spec/goldenSignals/processingDurationMetric", manifest.Spec.GoldenSignals.ProcessingDurationMetric, []string{"_seconds"})...)
	}
	for index, metric := range manifest.Spec.GoldenSignals.SaturationMetrics {
		path := fmt.Sprintf("/spec/goldenSignals/saturationMetrics/%d", index)
		issues = append(issues, metricConventionIssues(path, metric, []string{"_bytes", "_seconds", "_ratio", "_percent", "_count", "_total"})...)
	}
	return issues
}

func metricConventionIssues(path, metric string, suffixes []string) []Issue {
	if metric == "" {
		return nil
	}
	issues := make([]Issue, 0, 1)
	if !metricNamePattern.MatchString(metric) {
		issues = append(issues, Issue{Severity: SeverityError, Path: path, Message: "must be a lowercase Prometheus metric name"})
		return issues
	}
	for _, suffix := range suffixes {
		if strings.HasSuffix(metric, suffix) {
			return issues
		}
	}
	issues = append(issues, Issue{
		Severity: SeverityWarning,
		Path:     path,
		Message:  fmt.Sprintf("metric %q should end with one of: %s", metric, strings.Join(suffixes, ", ")),
	})
	return issues
}
