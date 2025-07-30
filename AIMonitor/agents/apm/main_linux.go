//go:build linux
// +build linux

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/elastic/go-elasticsearch/v8"
	"aimonitor-agents/common"
)

// LinuxAPMMonitor Linux版本的APM监控器
type LinuxAPMMonitor struct {
	*APMMonitor
	prometheusClient api.Client
	prometheusAPI    v1.API
	elasticsearchClient *elasticsearch.Client
	grafanaClient    *http.Client
	prometheusURL    string
	grafanaURL       string
	jaegerURL        string
	zipkinURL        string
	elasticAPMURL    string
	newRelicAPIKey   string
	skyWalkingURL    string
}

// NewLinuxAPMMonitor 创建Linux版本的APM监控器
func NewLinuxAPMMonitor(agent *common.Agent) *LinuxAPMMonitor {
	baseMonitor := NewAPMMonitor(agent)
	return &LinuxAPMMonitor{
		APMMonitor:      baseMonitor,
		prometheusURL:   "http://localhost:9090",
		grafanaURL:      "http://localhost:3000",
		jaegerURL:       "http://localhost:16686",
		zipkinURL:       "http://localhost:9411",
		elasticAPMURL:   "http://localhost:8200",
		newRelicAPIKey:  "your-new-relic-api-key",
		skyWalkingURL:   "http://localhost:8080",
		grafanaClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// initClients 初始化各种APM客户端
func (m *LinuxAPMMonitor) initClients() error {
	// 初始化Prometheus客户端
	if m.prometheusClient == nil {
		client, err := api.NewClient(api.Config{
			Address: m.prometheusURL,
		})
		if err != nil {
			return fmt.Errorf("failed to create Prometheus client: %v", err)
		}
		m.prometheusClient = client
		m.prometheusAPI = v1.NewAPI(client)
	}

	// 初始化Elasticsearch客户端
	if m.elasticsearchClient == nil {
		cfg := elasticsearch.Config{
			Addresses: []string{m.elasticAPMURL},
		}
		client, err := elasticsearch.NewClient(cfg)
		if err != nil {
			return fmt.Errorf("failed to create Elasticsearch client: %v", err)
		}
		m.elasticsearchClient = client
	}

	return nil
}

// makeHTTPRequest 发送HTTP请求
func (m *LinuxAPMMonitor) makeHTTPRequest(url string, headers map[string]string) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}

// getPrometheusMetrics 获取真实的Prometheus指标
func (m *LinuxAPMMonitor) getPrometheusMetrics() (map[string]interface{}, error) {
	ctx := context.Background()
	result := make(map[string]interface{})

	// 查询Prometheus服务器状态
	config, err := m.prometheusAPI.Config(ctx)
	if err == nil {
		result["prometheus_config_status"] = "loaded"
		result["prometheus_config_yaml_length"] = len(config.YAML)
	} else {
		result["prometheus_config_status"] = "error"
		result["prometheus_config_error"] = err.Error()
	}

	// 查询目标状态
	targets, err := m.prometheusAPI.Targets(ctx)
	if err == nil {
		activeTargets := 0
		droppedTargets := 0
		healthyTargets := 0
		unhealthyTargets := 0

		for _, target := range targets.Active {
			activeTargets++
			if target.Health == "up" {
				healthyTargets++
			} else {
				unhealthyTargets++
			}
		}
		droppedTargets = len(targets.Dropped)

		result["prometheus_active_targets"] = activeTargets
		result["prometheus_dropped_targets"] = droppedTargets
		result["prometheus_healthy_targets"] = healthyTargets
		result["prometheus_unhealthy_targets"] = unhealthyTargets
	}

	// 查询一些关键指标
	queries := map[string]string{
		"prometheus_tsdb_samples_total":     "prometheus_tsdb_samples_appended_total",
		"prometheus_tsdb_series_total":      "prometheus_tsdb_symbol_table_size_bytes",
		"prometheus_rule_evaluations_total": "prometheus_rule_evaluations_total",
		"prometheus_notifications_total":    "prometheus_notifications_total",
		"up_instances":                      "up",
	}

	for metricName, query := range queries {
		value, _, err := m.prometheusAPI.Query(ctx, query, time.Now())
		if err == nil {
			if vector, ok := value.(model.Vector); ok && len(vector) > 0 {
				if metricName == "up_instances" {
					upCount := 0
					for _, sample := range vector {
						if float64(sample.Value) == 1 {
							upCount++
						}
					}
					result[metricName] = upCount
				} else {
					result[metricName] = float64(vector[0].Value)
				}
			}
		}
	}

	// 查询存储使用情况
	storageQueries := map[string]string{
		"prometheus_tsdb_head_samples":       "prometheus_tsdb_head_samples",
		"prometheus_tsdb_head_series":        "prometheus_tsdb_head_series",
		"prometheus_tsdb_wal_size_bytes":     "prometheus_tsdb_wal_size_bytes",
		"prometheus_tsdb_head_chunks":        "prometheus_tsdb_head_chunks",
	}

	for metricName, query := range storageQueries {
		value, _, err := m.prometheusAPI.Query(ctx, query, time.Now())
		if err == nil {
			if vector, ok := value.(model.Vector); ok && len(vector) > 0 {
				result[metricName] = float64(vector[0].Value)
			}
		}
	}

	return result, nil
}

// getGrafanaMetrics 获取真实的Grafana指标
func (m *LinuxAPMMonitor) getGrafanaMetrics() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 获取Grafana健康状态
	healthURL := m.grafanaURL + "/api/health"
	healthData, err := m.makeHTTPRequest(healthURL, nil)
	if err == nil {
		var health map[string]interface{}
		if json.Unmarshal(healthData, &health) == nil {
			result["grafana_health_status"] = health["database"]
			result["grafana_version"] = health["version"]
		}
	} else {
		result["grafana_health_status"] = "error"
		result["grafana_health_error"] = err.Error()
	}

	// 获取数据源信息
	datasourcesURL := m.grafanaURL + "/api/datasources"
	datasourcesData, err := m.makeHTTPRequest(datasourcesURL, map[string]string{
		"Authorization": "Bearer your-grafana-api-key",
	})
	if err == nil {
		var datasources []map[string]interface{}
		if json.Unmarshal(datasourcesData, &datasources) == nil {
			result["grafana_total_datasources"] = len(datasources)
			
			// 统计数据源类型
			typeCount := make(map[string]int)
			for _, ds := range datasources {
				if dsType, ok := ds["type"].(string); ok {
					typeCount[dsType]++
				}
			}
			result["grafana_datasource_types"] = typeCount
		}
	}

	// 获取仪表板信息
	dashboardsURL := m.grafanaURL + "/api/search?type=dash-db"
	dashboardsData, err := m.makeHTTPRequest(dashboardsURL, map[string]string{
		"Authorization": "Bearer your-grafana-api-key",
	})
	if err == nil {
		var dashboards []map[string]interface{}
		if json.Unmarshal(dashboardsData, &dashboards) == nil {
			result["grafana_total_dashboards"] = len(dashboards)
			
			// 统计仪表板标签
			tagCount := make(map[string]int)
			for _, dashboard := range dashboards {
				if tags, ok := dashboard["tags"].([]interface{}); ok {
					for _, tag := range tags {
						if tagStr, ok := tag.(string); ok {
							tagCount[tagStr]++
						}
					}
				}
			}
			result["grafana_dashboard_tags"] = tagCount
		}
	}

	// 获取用户信息
	usersURL := m.grafanaURL + "/api/users"
	usersData, err := m.makeHTTPRequest(usersURL, map[string]string{
		"Authorization": "Bearer your-grafana-api-key",
	})
	if err == nil {
		var users []map[string]interface{}
		if json.Unmarshal(usersData, &users) == nil {
			result["grafana_total_users"] = len(users)
			
			// 统计活跃用户
			activeUsers := 0
			for _, user := range users {
				if isActive, ok := user["isActive"].(bool); ok && isActive {
					activeUsers++
				}
			}
			result["grafana_active_users"] = activeUsers
		}
	}

	return result, nil
}

// getJaegerMetrics 获取真实的Jaeger指标
func (m *LinuxAPMMonitor) getJaegerMetrics() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 获取Jaeger服务列表
	servicesURL := m.jaegerURL + "/api/services"
	servicesData, err := m.makeHTTPRequest(servicesURL, nil)
	if err == nil {
		var services map[string]interface{}
		if json.Unmarshal(servicesData, &services) == nil {
			if data, ok := services["data"].([]interface{}); ok {
				result["jaeger_total_services"] = len(data)
				result["jaeger_service_names"] = data
			}
		}
	} else {
		result["jaeger_services_error"] = err.Error()
	}

	// 获取操作信息（以第一个服务为例）
	if services, ok := result["jaeger_service_names"].([]interface{}); ok && len(services) > 0 {
		if serviceName, ok := services[0].(string); ok {
			operationsURL := fmt.Sprintf("%s/api/operations?service=%s", m.jaegerURL, serviceName)
			operationsData, err := m.makeHTTPRequest(operationsURL, nil)
			if err == nil {
				var operations map[string]interface{}
				if json.Unmarshal(operationsData, &operations) == nil {
					if data, ok := operations["data"].([]interface{}); ok {
						result["jaeger_sample_service_operations"] = len(data)
						result["jaeger_sample_service_name"] = serviceName
					}
				}
			}
		}
	}

	// 获取追踪信息
	tracesURL := fmt.Sprintf("%s/api/traces?limit=10&lookback=1h", m.jaegerURL)
	tracesData, err := m.makeHTTPRequest(tracesURL, nil)
	if err == nil {
		var traces map[string]interface{}
		if json.Unmarshal(tracesData, &traces) == nil {
			if data, ok := traces["data"].([]interface{}); ok {
				result["jaeger_recent_traces_count"] = len(data)
				
				// 分析追踪数据
				totalSpans := 0
				totalDuration := int64(0)
				serviceSet := make(map[string]bool)
				
				for _, trace := range data {
					if traceMap, ok := trace.(map[string]interface{}); ok {
						if spans, ok := traceMap["spans"].([]interface{}); ok {
							totalSpans += len(spans)
							
							for _, span := range spans {
								if spanMap, ok := span.(map[string]interface{}); ok {
									if duration, ok := spanMap["duration"].(float64); ok {
										totalDuration += int64(duration)
									}
									if process, ok := spanMap["process"].(map[string]interface{}); ok {
										if serviceName, ok := process["serviceName"].(string); ok {
											serviceSet[serviceName] = true
										}
									}
								}
							}
						}
					}
				}
				
				result["jaeger_total_spans_in_recent_traces"] = totalSpans
				result["jaeger_avg_trace_duration_microseconds"] = totalDuration / int64(len(data))
				result["jaeger_unique_services_in_traces"] = len(serviceSet)
			}
		}
	}

	return result, nil
}

// getZipkinMetrics 获取真实的Zipkin指标
func (m *LinuxAPMMonitor) getZipkinMetrics() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 获取Zipkin服务列表
	servicesURL := m.zipkinURL + "/api/v2/services"
	servicesData, err := m.makeHTTPRequest(servicesURL, nil)
	if err == nil {
		var services []string
		if json.Unmarshal(servicesData, &services) == nil {
			result["zipkin_total_services"] = len(services)
			result["zipkin_service_names"] = services
		}
	} else {
		result["zipkin_services_error"] = err.Error()
	}

	// 获取span名称（以第一个服务为例）
	if services, ok := result["zipkin_service_names"].([]string); ok && len(services) > 0 {
		serviceName := services[0]
		spanNamesURL := fmt.Sprintf("%s/api/v2/spans?serviceName=%s", m.zipkinURL, serviceName)
		spanNamesData, err := m.makeHTTPRequest(spanNamesURL, nil)
		if err == nil {
			var spanNames []string
			if json.Unmarshal(spanNamesData, &spanNames) == nil {
				result["zipkin_sample_service_span_names"] = len(spanNames)
				result["zipkin_sample_service_name"] = serviceName
			}
		}
	}

	// 获取最近的追踪
	endTs := time.Now().UnixNano() / 1000000 // 转换为毫秒
	startTs := endTs - 3600000              // 1小时前
	tracesURL := fmt.Sprintf("%s/api/v2/traces?endTs=%d&lookback=%d&limit=10", m.zipkinURL, endTs, 3600000)
	tracesData, err := m.makeHTTPRequest(tracesURL, nil)
	if err == nil {
		var traces [][]map[string]interface{}
		if json.Unmarshal(tracesData, &traces) == nil {
			result["zipkin_recent_traces_count"] = len(traces)
			
			// 分析追踪数据
			totalSpans := 0
			totalDuration := int64(0)
			serviceSet := make(map[string]bool)
			
			for _, trace := range traces {
				totalSpans += len(trace)
				
				for _, span := range trace {
					if duration, ok := span["duration"].(float64); ok {
						totalDuration += int64(duration)
					}
					if localEndpoint, ok := span["localEndpoint"].(map[string]interface{}); ok {
						if serviceName, ok := localEndpoint["serviceName"].(string); ok {
							serviceSet[serviceName] = true
						}
					}
				}
			}
			
			result["zipkin_total_spans_in_recent_traces"] = totalSpans
			if len(traces) > 0 {
				result["zipkin_avg_trace_duration_microseconds"] = totalDuration / int64(len(traces))
			}
			result["zipkin_unique_services_in_traces"] = len(serviceSet)
		}
	}

	return result, nil
}

// getElasticAPMMetrics 获取真实的Elastic APM指标
func (m *LinuxAPMMonitor) getElasticAPMMetrics() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 获取APM服务器信息
	infoURL := m.elasticAPMURL + "/"
	infoData, err := m.makeHTTPRequest(infoURL, nil)
	if err == nil {
		var info map[string]interface{}
		if json.Unmarshal(infoData, &info) == nil {
			if build, ok := info["build_date"]; ok {
				result["elastic_apm_build_date"] = build
			}
			if version, ok := info["version"]; ok {
				result["elastic_apm_version"] = version
			}
		}
	} else {
		result["elastic_apm_info_error"] = err.Error()
	}

	// 通过Elasticsearch查询APM数据
	if m.elasticsearchClient != nil {
		// 查询服务数量
		servicesQuery := map[string]interface{}{
			"aggs": map[string]interface{}{
				"services": map[string]interface{}{
					"terms": map[string]interface{}{
						"field": "service.name",
						"size":  100,
					},
				},
			},
			"size": 0,
			"query": map[string]interface{}{
				"range": map[string]interface{}{
					"@timestamp": map[string]interface{}{
						"gte": "now-1h",
					},
				},
			},
		}

		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(servicesQuery); err == nil {
			res, err := m.elasticsearchClient.Search(
				m.elasticsearchClient.Search.WithContext(context.Background()),
				m.elasticsearchClient.Search.WithIndex("apm-*"),
				m.elasticsearchClient.Search.WithBody(&buf),
			)
			if err == nil {
				defer res.Body.Close()
				if res.IsError() {
					result["elastic_apm_services_query_error"] = res.String()
				} else {
					var searchResult map[string]interface{}
					if err := json.NewDecoder(res.Body).Decode(&searchResult); err == nil {
						if aggs, ok := searchResult["aggregations"].(map[string]interface{}); ok {
							if services, ok := aggs["services"].(map[string]interface{}); ok {
								if buckets, ok := services["buckets"].([]interface{}); ok {
									result["elastic_apm_total_services"] = len(buckets)
									
									// 获取服务名称和文档数量
									serviceStats := make(map[string]interface{})
									for _, bucket := range buckets {
										if bucketMap, ok := bucket.(map[string]interface{}); ok {
											if key, ok := bucketMap["key"].(string); ok {
												if docCount, ok := bucketMap["doc_count"].(float64); ok {
													serviceStats[key] = int(docCount)
												}
											}
										}
									}
									result["elastic_apm_service_stats"] = serviceStats
								}
							}
						}
					}
				}
			}
		}

		// 查询错误率
		errorQuery := map[string]interface{}{
			"aggs": map[string]interface{}{
				"error_rate": map[string]interface{}{
					"filters": map[string]interface{}{
						"filters": map[string]interface{}{
							"errors": map[string]interface{}{
								"term": map[string]interface{}{
									"processor.event": "error",
								},
							},
							"transactions": map[string]interface{}{
								"term": map[string]interface{}{
									"processor.event": "transaction",
								},
							},
						},
					},
				},
			},
			"size": 0,
			"query": map[string]interface{}{
				"range": map[string]interface{}{
					"@timestamp": map[string]interface{}{
						"gte": "now-1h",
					},
				},
			},
		}

		buf.Reset()
		if err := json.NewEncoder(&buf).Encode(errorQuery); err == nil {
			res, err := m.elasticsearchClient.Search(
				m.elasticsearchClient.Search.WithContext(context.Background()),
				m.elasticsearchClient.Search.WithIndex("apm-*"),
				m.elasticsearchClient.Search.WithBody(&buf),
			)
			if err == nil {
				defer res.Body.Close()
				if !res.IsError() {
					var searchResult map[string]interface{}
					if err := json.NewDecoder(res.Body).Decode(&searchResult); err == nil {
						if aggs, ok := searchResult["aggregations"].(map[string]interface{}); ok {
							if errorRate, ok := aggs["error_rate"].(map[string]interface{}); ok {
								if buckets, ok := errorRate["buckets"].(map[string]interface{}); ok {
									errorCount := 0
									transactionCount := 0
									
									if errors, ok := buckets["errors"].(map[string]interface{}); ok {
										if docCount, ok := errors["doc_count"].(float64); ok {
											errorCount = int(docCount)
										}
									}
									if transactions, ok := buckets["transactions"].(map[string]interface{}); ok {
										if docCount, ok := transactions["doc_count"].(float64); ok {
											transactionCount = int(docCount)
										}
									}
									
									result["elastic_apm_error_count"] = errorCount
									result["elastic_apm_transaction_count"] = transactionCount
									if transactionCount > 0 {
										result["elastic_apm_error_rate_percent"] = float64(errorCount) / float64(transactionCount) * 100
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return result, nil
}

// getNewRelicMetrics 获取真实的New Relic指标
func (m *LinuxAPMMonitor) getNewRelicMetrics() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// New Relic API需要API密钥
	if m.newRelicAPIKey == "" || m.newRelicAPIKey == "your-new-relic-api-key" {
		result["new_relic_error"] = "API key not configured"
		return result, nil
	}

	// 获取应用程序列表
	appsURL := "https://api.newrelic.com/v2/applications.json"
	appsData, err := m.makeHTTPRequest(appsURL, map[string]string{
		"X-Api-Key": m.newRelicAPIKey,
	})
	if err == nil {
		var apps map[string]interface{}
		if json.Unmarshal(appsData, &apps) == nil {
			if applications, ok := apps["applications"].([]interface{}); ok {
				result["new_relic_total_applications"] = len(applications)
				
				// 统计应用状态
				healthyApps := 0
				warningApps := 0
				criticalApps := 0
				
				for _, app := range applications {
					if appMap, ok := app.(map[string]interface{}); ok {
						if healthStatus, ok := appMap["health_status"].(string); ok {
							switch healthStatus {
							case "green":
								healthyApps++
							case "yellow":
								warningApps++
							case "red":
								criticalApps++
							}
						}
					}
				}
				
				result["new_relic_healthy_applications"] = healthyApps
				result["new_relic_warning_applications"] = warningApps
				result["new_relic_critical_applications"] = criticalApps
			}
		}
	} else {
		result["new_relic_applications_error"] = err.Error()
	}

	// 获取服务器列表
	serversURL := "https://api.newrelic.com/v2/servers.json"
	serversData, err := m.makeHTTPRequest(serversURL, map[string]string{
		"X-Api-Key": m.newRelicAPIKey,
	})
	if err == nil {
		var servers map[string]interface{}
		if json.Unmarshal(serversData, &servers) == nil {
			if serverList, ok := servers["servers"].([]interface{}); ok {
				result["new_relic_total_servers"] = len(serverList)
				
				// 统计服务器状态
				reportingServers := 0
				for _, server := range serverList {
					if serverMap, ok := server.(map[string]interface{}); ok {
						if reporting, ok := serverMap["reporting"].(bool); ok && reporting {
							reportingServers++
						}
					}
				}
				
				result["new_relic_reporting_servers"] = reportingServers
			}
		}
	}

	return result, nil
}

// getSkyWalkingMetrics 获取真实的SkyWalking指标
func (m *LinuxAPMMonitor) getSkyWalkingMetrics() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 获取SkyWalking服务列表
	servicesURL := m.skyWalkingURL + "/graphql"
	query := `{
		getAllServices(duration: {start: "2023-12-15 09", end: "2023-12-15 10", step: HOUR}) {
			id
			name
			group
			layers
		}
	}`

	reqBody := map[string]string{"query": query}
	reqData, _ := json.Marshal(reqBody)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", servicesURL, bytes.NewBuffer(reqData))
	if err == nil {
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == 200 {
				var result_data map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&result_data); err == nil {
					if data, ok := result_data["data"].(map[string]interface{}); ok {
						if services, ok := data["getAllServices"].([]interface{}); ok {
							result["skywalking_total_services"] = len(services)
							
							// 统计服务层级
							layerCount := make(map[string]int)
							for _, service := range services {
								if serviceMap, ok := service.(map[string]interface{}); ok {
									if layers, ok := serviceMap["layers"].([]interface{}); ok {
										for _, layer := range layers {
											if layerStr, ok := layer.(string); ok {
												layerCount[layerStr]++
											}
										}
									}
								}
							}
							result["skywalking_service_layers"] = layerCount
						}
					}
				}
			} else {
				result["skywalking_services_error"] = fmt.Sprintf("HTTP %d", resp.StatusCode)
			}
		} else {
			result["skywalking_services_error"] = err.Error()
		}
	} else {
		result["skywalking_services_error"] = err.Error()
	}

	// 获取拓扑信息
	topologyQuery := `{
		getGlobalTopology(duration: {start: "2023-12-15 09", end: "2023-12-15 10", step: HOUR}) {
			nodes {
				id
				name
				type
				isReal
			}
			calls {
				source
				target
				id
			}
		}
	}`

	topologyReqBody := map[string]string{"query": topologyQuery}
	topologyReqData, _ := json.Marshal(topologyReqBody)

	topologyReq, err := http.NewRequest("POST", servicesURL, bytes.NewBuffer(topologyReqData))
	if err == nil {
		topologyReq.Header.Set("Content-Type", "application/json")
		topologyResp, err := client.Do(topologyReq)
		if err == nil {
			defer topologyResp.Body.Close()
			if topologyResp.StatusCode == 200 {
				var topologyResult map[string]interface{}
				if err := json.NewDecoder(topologyResp.Body).Decode(&topologyResult); err == nil {
					if data, ok := topologyResult["data"].(map[string]interface{}); ok {
						if topology, ok := data["getGlobalTopology"].(map[string]interface{}); ok {
							if nodes, ok := topology["nodes"].([]interface{}); ok {
								result["skywalking_topology_nodes"] = len(nodes)
							}
							if calls, ok := topology["calls"].([]interface{}); ok {
								result["skywalking_topology_calls"] = len(calls)
							}
						}
					}
				}
			}
		}
	}

	return result, nil
}

// 重写collectMetrics方法以使用Linux特定的实现
func (m *LinuxAPMMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// 初始化客户端
	if err := m.initClients(); err != nil {
		m.agent.Logger.Error("Failed to initialize APM clients: %v", err)
		return m.APMMonitor.collectMetrics() // 回退到模拟数据
	}

	// 添加连接信息
	metrics["prometheus_url"] = m.prometheusURL
	metrics["grafana_url"] = m.grafanaURL
	metrics["jaeger_url"] = m.jaegerURL
	metrics["zipkin_url"] = m.zipkinURL
	metrics["elastic_apm_url"] = m.elasticAPMURL
	metrics["skywalking_url"] = m.skyWalkingURL
	metrics["collection_time"] = time.Now().Format(time.RFC3339)

	// 使用Linux特定的方法收集指标
	if prometheusMetrics, err := m.getPrometheusMetrics(); err == nil {
		for k, v := range prometheusMetrics {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get Prometheus metrics: %v", err)
	}

	if grafanaMetrics, err := m.getGrafanaMetrics(); err == nil {
		for k, v := range grafanaMetrics {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get Grafana metrics: %v", err)
	}

	if jaegerMetrics, err := m.getJaegerMetrics(); err == nil {
		for k, v := range jaegerMetrics {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get Jaeger metrics: %v", err)
	}

	if zipkinMetrics, err := m.getZipkinMetrics(); err == nil {
		for k, v := range zipkinMetrics {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get Zipkin metrics: %v", err)
	}

	if elasticAPMMetrics, err := m.getElasticAPMMetrics(); err == nil {
		for k, v := range elasticAPMMetrics {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get Elastic APM metrics: %v", err)
	}

	if newRelicMetrics, err := m.getNewRelicMetrics(); err == nil {
		for k, v := range newRelicMetrics {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get New Relic metrics: %v", err)
	}

	if skyWalkingMetrics, err := m.getSkyWalkingMetrics(); err == nil {
		for k, v := range skyWalkingMetrics {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get SkyWalking metrics: %v", err)
	}

	// 如果所有API调用都失败，回退到模拟数据
	if len(metrics) <= 7 { // 只有基本连接信息
		m.agent.Logger.Warn("All APM API calls failed, falling back to simulated data")
		return m.APMMonitor.collectMetrics()
	}

	return metrics
}