# obs-manifest-schema

Servis takımlarının kendi reposunda tuttuğu `observability.yaml` sözleşmesinin resmi JSON Schema'ları, offline doğrulayıcısı ve migration aracı.

## Hızlı başlangıç

```bash
go install github.com/acme/obs-manifest-schema/cmd/obsctl@latest
obsctl validate observability.yaml
obsctl migrate v1 v2 -o observability-v2.yaml observability.yaml
```

GitHub Action kullanımı:

```yaml
name: Observability manifest
on: [pull_request]
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
      - uses: acme/obs-manifest-schema@v1
        with:
          path: observability.yaml
          warnings-as-errors: "false"
```

Action şemaları ve CLI'ı kendi checkout'undan derler; doğrulama sırasında harici schema servisine ihtiyaç duymaz. `validate` geçersiz manifestte veya `--warnings-as-errors` aktifken warning bulunduğunda exit code `1`, yanlış komut kullanımında `2` döndürür.

v1 manifestleri deprecation warning ürettiği için migration tamamlanana kadar `warnings-as-errors` kapalı tutulur. v2'ye geçildikten sonra naming warning'lerini zorunlu politika yapmak için değer `"true"` yapılabilir.

## Şema kimliği

| Manifest API | JSON Schema `$id` | Schema sürümü | Durum |
|---|---|---:|---|
| `obs.acme.com/v1` | `https://schemas.obs.acme.com/service-observability/v1/schema.json` | `1.0.0` | Deprecated, doğrulanmaya devam eder |
| `obs.acme.com/v2` | `https://schemas.obs.acme.com/service-observability/v2/schema.json` | `2.0.0` | Current |

Her iki dosya Draft 2020-12 kullanır. Resmi kaynaklar [v1 schema](schemas/v1/service-observability.schema.json) ve [v2 schema](schemas/v2/service-observability.schema.json) dosyalarıdır.

## v1 referansı

| Alan | Tür | Zorunlu | Kural |
|---|---|---:|---|
| `apiVersion` | string | Evet | `obs.acme.com/v1` |
| `kind` | string | Evet | `ServiceObservability` |
| `metadata.name` | string | Evet | DNS-benzeri lowercase servis adı |
| `metadata.owner` | string | Evet | Takım slug'ı |
| `metadata.slack_channel` | string | Evet | `#` ile başlayan lowercase kanal |
| `metadata.pagerduty_service` | string | Evet | PagerDuty service adı |
| `spec.serviceType` | enum | Evet | `http`, `grpc`, `worker`, `batch`, `database` |
| `spec.environment` | enum | Evet | `dev`, `staging`, `prod` |
| `spec.slo.availability_target` | number | Evet | `0.9..0.99999` |
| `spec.slo.latency_target_ms` | integer | Evet | Pozitif, en fazla 24 saat |
| `spec.slo.latency_percentile` | integer | Evet | Semantik olarak yalnız `50`, `90`, `95`, `99` |
| `spec.goldenSignals.trafficMetric` | metric | Evet | Counter, `_total` |
| `spec.goldenSignals.errorMetric` | metric | Evet | Counter, `_total` |
| `spec.goldenSignals.latencyMetric` | metric | Koşullu | `http`, `grpc`, `database`; `_seconds` |
| `spec.goldenSignals.processingDurationMetric` | metric | Koşullu | `worker`, `batch`; `_seconds` |
| `spec.goldenSignals.saturationMetrics` | metric[] | Evet | En az bir saturation metriği |
| `spec.dashboards.template` | string | Evet | Dashboard catalog template slug'ı |
| `spec.alerting.burnRateAlerts` | boolean | Evet | SLO multi-window burn-rate alert üretimi |
| `spec.alerting.extraRules` | object[] | Evet | `name`, `expr`, `for`, `severity`, `summary` |

Bilinmeyen alanlar reddedilir. Bir dosyada yalnız tek YAML document bulunabilir. Ayrıntılı metrik kuralları [naming-convention.md](naming-convention.md) içindedir.

## Tam örnekler

### HTTP

```yaml
apiVersion: obs.acme.com/v1
kind: ServiceObservability
metadata:
  name: checkout-api
  owner: team-checkout
  slack_channel: "#checkout-alerts"
  pagerduty_service: checkout-production
spec:
  serviceType: http
  environment: prod
  slo:
    availability_target: 0.999
    latency_target_ms: 300
    latency_percentile: 99
  goldenSignals:
    trafficMetric: http_server_requests_total
    errorMetric: http_server_request_errors_total
    latencyMetric: http_server_request_duration_seconds
    saturationMetrics:
      - http_server_active_requests_count
      - process_resident_memory_bytes
  dashboards:
    template: golden-signals/http-v1
  alerting:
    burnRateAlerts: true
    extraRules:
      - name: CheckoutQueueBacklogHigh
        expr: checkout_pending_orders_count > 500
        for: 15m
        severity: warning
        summary: Checkout pending-order backlog is elevated
```

### Worker

```yaml
apiVersion: obs.acme.com/v1
kind: ServiceObservability
metadata:
  name: invoice-worker
  owner: team-billing
  slack_channel: "#billing-alerts"
  pagerduty_service: billing-production
spec:
  serviceType: worker
  environment: prod
  slo:
    availability_target: 0.995
    latency_target_ms: 30000
    latency_percentile: 95
  goldenSignals:
    trafficMetric: invoice_jobs_processed_total
    errorMetric: invoice_jobs_failed_total
    processingDurationMetric: invoice_job_processing_duration_seconds
    saturationMetrics:
      - invoice_queue_depth_count
      - process_resident_memory_bytes
  dashboards:
    template: golden-signals/worker-v1
  alerting:
    burnRateAlerts: true
    extraRules:
      - name: InvoiceQueueStalled
        expr: rate(invoice_jobs_processed_total[10m]) == 0 and invoice_queue_depth_count > 0
        for: 10m
        severity: critical
        summary: Invoice queue has work but no successful processing
```

### Batch

```yaml
apiVersion: obs.acme.com/v1
kind: ServiceObservability
metadata:
  name: nightly-ledger-export
  owner: team-finance-platform
  slack_channel: "#finance-platform-alerts"
  pagerduty_service: finance-batch-production
spec:
  serviceType: batch
  environment: prod
  slo:
    availability_target: 0.99
    latency_target_ms: 3600000
    latency_percentile: 90
  goldenSignals:
    trafficMetric: ledger_export_runs_total
    errorMetric: ledger_export_failures_total
    processingDurationMetric: ledger_export_duration_seconds
    saturationMetrics:
      - ledger_export_pending_partitions_count
      - ledger_export_buffer_bytes
  dashboards:
    template: golden-signals/batch-v1
  alerting:
    burnRateAlerts: true
    extraRules:
      - name: LedgerExportMissed
        expr: time() - ledger_export_last_success_seconds > 90000
        for: 5m
        severity: critical
        summary: Nightly ledger export has missed its completion window
```

### Database

```yaml
apiVersion: obs.acme.com/v1
kind: ServiceObservability
metadata:
  name: customer-postgres
  owner: team-data-platform
  slack_channel: "#data-platform-alerts"
  pagerduty_service: customer-database-production
spec:
  serviceType: database
  environment: prod
  slo:
    availability_target: 0.9999
    latency_target_ms: 100
    latency_percentile: 99
  goldenSignals:
    trafficMetric: database_client_operations_total
    errorMetric: database_client_errors_total
    latencyMetric: database_client_operation_duration_seconds
    saturationMetrics:
      - database_connections_usage_ratio
      - database_storage_used_bytes
  dashboards:
    template: golden-signals/database-v1
  alerting:
    burnRateAlerts: true
    extraRules:
      - name: DatabaseConnectionsNearLimit
        expr: database_connections_usage_ratio > 0.85
        for: 10m
        severity: warning
        summary: Database connection pool is approaching its limit
```

Çalıştırılabilir kopyalar [testdata/valid](testdata/valid) dizinindedir.

## Şema evrimi

- Aynı major sürümde yalnız backward-compatible değişiklik yapılır: yeni opsiyonel alan, enum davranışını daraltmayan açıklama veya ek warning. Var olan required alan kaldırılmaz, alan tipi değiştirilmez ve enum daraltılmaz.
- CLI yayımlanmış bütün major şemaları binary içinde taşır. Bu nedenle yeni CLI v1 dosyasını doğrular; v1 için migration öneren deprecation warning verir ancak varsayılan olarak build'i kırmaz.
- Breaking değişiklik yeni `apiVersion` ve yeni `$id` gerektirir. Eski şema dosyası immutable kalır.
- Her major kenarı için çalışan ve test edilen migration fonksiyonu gerekir. v1→v2 dönüşümü metadata contact alanlarını `spec.ownership.team`, `slackChannel`, `pagerDutyService` alanlarına taşır.
- Migration deterministiktir, kaynak dosyayı yerinde ezmez ve çıktıyı hedef schema ile yeniden doğrulamadan yazmaz.

```bash
obsctl migrate v1 v2 -o observability-v2.yaml observability.yaml
obsctl validate observability-v2.yaml
```

## Geliştirme

```bash
make check
./bin/obsctl validate testdata/valid/http.yaml
```

Testler schema kimliklerini, dört servis tipini, semantik kuralları, warning politikasını, hatalı fixture'ları ve migration sonucunu table-driven Go testleriyle doğrular.
