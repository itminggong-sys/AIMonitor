module aimonitor-vmware-agent

go 1.21

require (
	aimonitor-agents/common v0.0.0
	github.com/vmware/govmomi v0.33.1
	github.com/google/uuid v1.4.0
	gopkg.in/yaml.v3 v3.0.1
)

replace aimonitor-agents/common => ../common