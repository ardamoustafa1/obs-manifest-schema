package manifest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidFixtures(t *testing.T) {
	t.Parallel()
	validator := mustValidator(t)
	tests := []struct {
		name        string
		serviceType string
	}{
		{"http.yaml", "http"},
		{"worker.yaml", "worker"},
		{"batch.yaml", "batch"},
		{"database.yaml", "database"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			data := readFixture(t, "valid", test.name)
			report, parsed := validator.Validate(data)
			if !report.Valid() {
				t.Fatalf("expected valid manifest, got issues: %v", report.Issues)
			}
			if parsed == nil || parsed.Spec.ServiceType != test.serviceType {
				t.Fatalf("expected serviceType %q, got %#v", test.serviceType, parsed)
			}
			if !report.HasWarnings() {
				t.Fatal("v1 manifest should carry a deprecation warning")
			}
		})
	}
}

func TestInvalidFixtures(t *testing.T) {
	t.Parallel()
	validator := mustValidator(t)
	tests := []struct {
		name     string
		wantText string
	}{
		{"percentile.yaml", "must be one of 50, 90, 95, or 99"},
		{"worker-latency.yaml", "processingDurationMetric"},
		{"unknown-field.yaml", "additional properties"},
		{"metric-name.yaml", "pattern"},
		{"multiple-documents.yaml", "exactly one YAML document"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			report, _ := validator.Validate(readFixture(t, "invalid", test.name))
			if report.Valid() {
				t.Fatal("expected validation to fail")
			}
			if !strings.Contains(issuesText(report.Issues), test.wantText) {
				t.Fatalf("issues %q do not contain %q", issuesText(report.Issues), test.wantText)
			}
		})
	}
}

func TestNamingWarningDoesNotInvalidateManifest(t *testing.T) {
	t.Parallel()
	validator := mustValidator(t)
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "warning-metric-suffix.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	report, _ := validator.Validate(data)
	if !report.Valid() {
		t.Fatalf("naming warnings must not invalidate by default: %v", report.Issues)
	}
	if !report.HasWarnings() || !strings.Contains(issuesText(report.Issues), "_total") {
		t.Fatalf("expected a suffix warning, got: %v", report.Issues)
	}
}

func TestUnsupportedVersion(t *testing.T) {
	t.Parallel()
	validator := mustValidator(t)
	data := []byte("apiVersion: obs.acme.com/v9\nkind: ServiceObservability\nmetadata: {}\nspec: {}\n")
	report, _ := validator.Validate(data)
	if report.Valid() || !strings.Contains(issuesText(report.Issues), "unsupported apiVersion") {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func mustValidator(t *testing.T) *Validator {
	t.Helper()
	validator, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}
	return validator
}

func readFixture(t *testing.T, category, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", category, name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func issuesText(issues []Issue) string {
	var builder strings.Builder
	for _, issue := range issues {
		builder.WriteString(issue.String())
		builder.WriteByte('\n')
	}
	return builder.String()
}
