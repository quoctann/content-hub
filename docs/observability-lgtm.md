# Content Hub observability on k3s LGTM

This document is the handoff for the separate `content-gitops` repository.
It assumes a single-node k3s cluster, an already installed persistent LGTM
deployment, 7-14 day retention, and Slack notifications.

## Backend configuration

The backend exposes Prometheus metrics at `GET /metrics` and sends traces via
OTLP/gRPC. Metrics listen only on the dedicated in-cluster port `9090`; set
these environment variables on the backend Deployment:

```yaml
env:
  - name: APP_ENV
    value: prod
  - name: SERVICE_VERSION
    value: "<immutable-image-sha>"
  - name: OTEL_SERVICE_NAME
    value: content-hub
  - name: OTEL_EXPORTER_OTLP_ENDPOINT
    value: "<lgtm-otlp-service>.observability.svc.cluster.local:4317"
  - name: OTEL_TRACES_SAMPLER_ARG
    value: "0.10"
  - name: OTEL_EXPORTER_OTLP_TLS
    value: "false"
  - name: METRICS_PORT
    value: "9090"
```

`OTEL_EXPORTER_OTLP_ENDPOINT` is deliberately host:port without `http://`.
The backend uses insecure gRPC only because traffic remains inside the cluster.
Set `OTEL_EXPORTER_OTLP_TLS=true` when the collector requires TLS. An empty
endpoint disables trace export but preserves W3C
`traceparent` propagation.

Start with `0.10` sampling. Temporarily use `1.0` only while validating the
deployment. Do not use a `VITE_*` variable for any backend observability
setting or secret.

## Workload and Service

Use stable, low-cardinality labels consistently on the Deployment, Pod
template, and Service:

```yaml
metadata:
  labels:
    app.kubernetes.io/name: content-hub
    app.kubernetes.io/component: backend
    app.kubernetes.io/part-of: content-hub
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: content-hub
      app.kubernetes.io/component: backend
```

The Service must select the same Pod labels. Expose both named ports; do not
create a public Ingress route for the `metrics` port.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: content-hub
  namespace: content
  labels:
    app.kubernetes.io/name: content-hub
    app.kubernetes.io/component: backend
spec:
  selector:
    app.kubernetes.io/name: content-hub
    app.kubernetes.io/component: backend
  ports:
    - name: http
      port: 8080
      targetPort: 8080
    - name: metrics
      port: 9090
      targetPort: 9090
```

Configure probes as follows:

```yaml
startupProbe:
  httpGet: {path: /health/ready, port: http}
  periodSeconds: 5
  failureThreshold: 12
livenessProbe:
  httpGet: {path: /health/live, port: http}
  periodSeconds: 10
  failureThreshold: 3
readinessProbe:
  httpGet: {path: /health/ready, port: http}
  periodSeconds: 5
  failureThreshold: 2
```

## Prometheus scraping

If LGTM uses Prometheus Operator, apply this `ServiceMonitor`. Replace
`content` only when the application namespace differs.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: content-hub
  namespace: observability
  labels:
    release: lgtm
spec:
  namespaceSelector:
    matchNames: [content]
  selector:
    matchLabels:
      app.kubernetes.io/name: content-hub
      app.kubernetes.io/component: backend
  endpoints:
    - port: metrics
      path: /metrics
      interval: 30s
      scrapeTimeout: 10s
```

The `release: lgtm` label is chart-specific. Match it to the label selector
configured on the existing Prometheus instance. For plain Prometheus, use
Kubernetes Service discovery with the same Service labels and named `metrics`
port instead.

Important metric names and labels:

```text
content_hub_http_requests_total{method,route,status}
content_hub_http_request_duration_seconds{method,route,status}
content_hub_http_requests_in_flight{method,route}
content_hub_database_pool_acquired_connections
content_hub_database_pool_idle_connections
content_hub_database_pool_total_connections
content_hub_database_pool_max_connections
```

`route` is the Gin route template, for example `/contents/:id`, and unmatched
requests are `unmatched`; raw paths and query strings are never metric labels.

## Alloy and Loki logs

Deploy Alloy as a DaemonSet with access to k3s container logs. Discover Pods
and keep only content workload logs using this selector:

```text
namespace=content, app.kubernetes.io/name=content-hub
```

Parse CRI then JSON log lines. Send Loki labels only for:

```text
namespace, app, container, environment, level
```

Keep `trace_id`, `span_id`, `request_id`, `path`, `client_ip`, `user_agent`,
and `error` as JSON fields or structured metadata, never Loki index labels.
The backend log fields are `timestamp`, `level`, `service`, `env`, `msg`,
`trace_id`, `span_id`, and `request_id` where applicable.

Provision a Grafana Loki derived field that extracts `trace_id` and opens the
Tempo datasource. Configure Tempo's `tracesToLogs` link to query Loki with the
same trace ID as a parsed JSON field, for example:

```logql
{app="content-hub"} | json | service="content-hub" | trace_id="${__trace.traceId}"
```

## Grafana and Slack

In Grafana, verify or provision these data sources:

```text
Prometheus/Mimir: <lgtm-prometheus-service>:9090
Loki:             <lgtm-loki-service>:3100
Tempo:            <lgtm-tempo-service>:3200
```

Use Grafana provisioning files or Grafana-managed resources in GitOps, not
manual UI configuration. Store the Slack webhook in a Kubernetes Secret and
reference it from the Grafana contact point. Configure notification policies:

```text
severity=critical -> Slack immediately
severity=warning  -> Slack after 10 minutes
```

Initial alerts:

```text
BackendDown:                 up{service="content-hub"} == 0 for 5m
BackendHigh5xxRate:          5xx ratio > 5% for 5m, with request-volume guard
BackendHighP95Latency:       p95 > 1 second for 10m
BackendPanic:                Loki count of msg="panic_recovered" > 0 for 1m
PodCrashLooping:             restart increase over 10m
PostgresUnavailable:         postgres exporter or backend readiness failure
PersistentVolumeNearFull:    >80% warning, >90% critical
CertificateExpiringSoon:     less than 14 days via blackbox exporter
```

Add `postgres_exporter` with a dedicated least-privileged database user and
blackbox probes for the public frontend, `/api/health`, and TLS certificate.
Do not scrape database credentials, JWT values, request bodies, SQL arguments,
or authentication usernames into logs, metrics, traces, or dashboards.

Apply a NetworkPolicy that permits TCP/9090 ingress only from the Prometheus
or Alloy scraper Pods in the `observability` namespace. The frontend and
Ingress need only TCP/8080. This is required even though the metrics port is
not attached to public Ingress.
