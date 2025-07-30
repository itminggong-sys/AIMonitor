module aimonitor-apm-agent

go 1.21

require (
	aimonitor-agents/common v0.0.0
	github.com/prometheus/client_golang v1.17.0
	github.com/prometheus/common v0.45.0
	github.com/grafana/grafana-api-golang-client v0.23.0
	github.com/jaegertracing/jaeger v1.50.0
	github.com/elastic/go-elasticsearch/v8 v8.11.1
	github.com/google/uuid v1.4.0
	gopkg.in/yaml.v3 v3.0.1
)

replace aimonitor-agents/common => ../common