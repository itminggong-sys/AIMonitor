module aimonitor-apache-agent

go 1.21

require (
	aimonitor-agents/common v0.0.0
	github.com/google/uuid v1.3.0
	gopkg.in/yaml.v3 v3.0.1
)

replace aimonitor-agents/common => ../common