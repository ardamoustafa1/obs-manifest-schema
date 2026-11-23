package manifest

import "testing"

func TestMetricConventionIssues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		metric   string
		suffixes []string
		severity Severity
		count    int
	}{
		{name: "counter", metric: "http_server_requests_total", suffixes: []string{"_total"}, count: 0},
		{name: "duration", metric: "http_server_duration_seconds", suffixes: []string{"_seconds"}, count: 0},
		{name: "bad suffix", metric: "http_server_requests", suffixes: []string{"_total"}, severity: SeverityWarning, count: 1},
		{name: "invalid syntax", metric: "HTTP-Requests", suffixes: []string{"_total"}, severity: SeverityError, count: 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			issues := metricConventionIssues("/metric", test.metric, test.suffixes)
			if len(issues) != test.count {
				t.Fatalf("got %d issues, want %d: %#v", len(issues), test.count, issues)
			}
			if test.count > 0 && issues[0].Severity != test.severity {
				t.Fatalf("got severity %s, want %s", issues[0].Severity, test.severity)
			}
		})
	}
}
