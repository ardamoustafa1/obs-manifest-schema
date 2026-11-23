# Golden signal metric naming convention

`observability.yaml` alanları Prometheus tarafından sorgulanan, OpenTelemetry instrument'larından export edilmiş metrik ailelerinin adlarını taşır. Kaynak instrument adları OpenTelemetry semantic conventions'a uygun noktalı biçimde olabilir (`http.server.request.duration` gibi); collector/Prometheus exporter sonrasında manifestte kullanılan ad lowercase snake_case ve birim suffix'li olmalıdır (`http_server_request_duration_seconds`).

## Genel kurallar

1. Adlar yalnız `[a-z0-9_:]` karakterlerini kullanır; küçük harfle veya `:` ile başlar.
2. Birim adın sonunda ve tekil yazılır: `_seconds`, `_bytes`, `_ratio`, `_percent`.
3. Counter aileleri `_total` ile biter. Histogram için `_bucket`, `_sum` veya `_count` serisini değil, aile adını yazın.
4. Base unit kullanın: süre için saniye, boyut için byte, oran için `0..1` arası ratio. `latency_target_ms` insan-okunur SLO girdisidir; latency metriğinin birimi yine saniyedir.
5. Servis namespace'i ve sinyal açık olmalıdır. `requests_total` yerine `checkout_http_server_requests_total` gibi çakışmayan bir ad tercih edin.
6. User ID, request ID, e-posta, ham URL veya sınırsız cardinality üreten değerleri label yapmayın. Route template, HTTP method, status code, RPC service/method ve kontrollü queue adı kullanılabilir.
7. OpenTelemetry standardında karşılığı olan sinyallerde aynı anlamı ve birimi koruyun. Uygulama içinde farklı bir instrument adı kullanılıyorsa collector transform kuralı ile aşağıdaki Prometheus sözleşmesine dönüştürün.

## Alan bazında kurallar

| Alan | Tür | Zorunlu suffix | Örnek |
|---|---|---|---|
| `trafficMetric` | Monotonic counter | `_total` | `http_server_requests_total` |
| `errorMetric` | Monotonic counter | `_total` | `http_server_request_errors_total` |
| `latencyMetric` | Histogram ailesi | `_seconds` | `http_server_request_duration_seconds` |
| `processingDurationMetric` | Histogram ailesi | `_seconds` | `invoice_job_processing_duration_seconds` |
| `saturationMetrics[]` | Gauge/counter | `_bytes`, `_seconds`, `_ratio`, `_percent`, `_count`, `_total` | `process_resident_memory_bytes` |

### OpenTelemetry eşlemeleri

| OpenTelemetry anlamı | Manifestte önerilen export adı |
|---|---|
| `http.server.request.duration`, unit `s` | `http_server_request_duration_seconds` |
| `rpc.server.duration`, unit `s` | `rpc_server_duration_seconds` |
| `db.client.operation.duration`, unit `s` | `database_client_operation_duration_seconds` |
| `process.memory.usage`, unit `By` | `process_memory_usage_bytes` |
| Job tüketim sayacı | `<domain>_jobs_processed_total` |

CLI, karakter seti ihlalini schema error olarak reddeder. Anlamlı fakat suffix standardına uymayan bir ad warning üretir; merkezi policy gereken repolarda Action'ın `warnings-as-errors: "true"` girdisi bu warning'i PR hatasına çevirir.

