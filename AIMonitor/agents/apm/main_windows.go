//go:build windows
// +build windows

package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"aimonitor-agents/common"
)

// WindowsAPMMonitor Windows版本的APM监控器
type WindowsAPMMonitor struct {
	*APMMonitor
	httpClient *http.Client
}

// NewWindowsAPMMonitor 创建Windows版本的APM监控器
func NewWindowsAPMMonitor(agent *common.Agent) *WindowsAPMMonitor {
	baseMonitor := NewAPMMonitor(agent)
	
	// 创建HTTP客户端
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	
	return &WindowsAPMMonitor{
		APMMonitor: baseMonitor,
		httpClient: httpClient,
	}
}

// Windows API 结构体和常量
type PROCESS_MEMORY_COUNTERS struct {
	Cb                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
}

type FILETIME struct {
	DwLowDateTime  uint32
	DwHighDateTime uint32
}

var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	psapi                   = syscall.NewLazyDLL("psapi.dll")
	getCurrentProcess       = kernel32.NewProc("GetCurrentProcess")
	getProcessMemoryInfo    = psapi.NewProc("GetProcessMemoryInfo")
	getProcessTimes         = kernel32.NewProc("GetProcessTimes")
	openProcess             = kernel32.NewProc("OpenProcess")
	closeHandle             = kernel32.NewProc("CloseHandle")
	getProcessImageFileName = psapi.NewProc("GetProcessImageFileNameW")
)

const (
	PROCESS_QUERY_INFORMATION = 0x0400
	PROCESS_VM_READ           = 0x0010
)

// sendHTTPRequest 发送HTTP请求
func (m *WindowsAPMMonitor) sendHTTPRequest(url string, headers map[string]string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	// 添加请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		// 如果不是JSON格式，返回原始文本
		result = map[string]interface{}{
			"raw_response": string(body),
			"status_code":  resp.StatusCode,
		}
	}
	
	result["status_code"] = resp.StatusCode
	return result, nil
}

// getAPMProcessInfo 获取APM相关进程信息
func (m *WindowsAPMMonitor) getAPMProcessInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// APM相关进程名称
	apmProcesses := []string{
		"prometheus.exe",
		"grafana-server.exe",
		"jaeger.exe",
		"zipkin.exe",
		"elasticsearch.exe",
		"kibana.exe",
		"java.exe", // 可能运行APM工具
		"node.exe", // 可能运行APM工具
	}
	
	allProcesses := make([]map[string]interface{}, 0)
	totalMemory := int64(0)
	
	for _, processName := range apmProcesses {
		cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s", processName), "/FO", "CSV")
		output, err := cmd.Output()
		if err != nil {
			continue
		}
		
		lines := strings.Split(string(output), "\n")
		for i, line := range lines {
			if i == 0 || strings.TrimSpace(line) == "" {
				continue // 跳过标题行和空行
			}
			
			// 解析CSV格式的输出
			fields := strings.Split(line, ",")
			if len(fields) >= 5 {
				imageName := strings.Trim(fields[0], `"`)
				pidStr := strings.Trim(fields[1], `"`)
				memUsageStr := strings.Trim(fields[4], `"`)
				
				// 检查是否是APM相关进程
				if m.isAPMProcess(imageName, pidStr) {
					pid, _ := strconv.Atoi(pidStr)
					memUsage := m.parseMemoryUsage(memUsageStr)
					totalMemory += memUsage
					
					processInfo := map[string]interface{}{
						"image_name":   imageName,
						"pid":          pid,
						"memory_usage": memUsage,
					}
					
					// 获取更详细的进程信息
					if detailedInfo, err := m.getDetailedProcessInfo(pid); err == nil {
						for k, v := range detailedInfo {
							processInfo[k] = v
						}
					}
					
					// 识别APM工具类型
					processInfo["apm_tool"] = m.identifyAPMTool(imageName, pidStr)
					
					allProcesses = append(allProcesses, processInfo)
				}
			}
		}
	}
	
	result["apm_processes_running"] = len(allProcesses) > 0
	result["process_count"] = len(allProcesses)
	result["processes"] = allProcesses
	result["total_memory_usage"] = totalMemory
	result["total_memory_usage_mb"] = totalMemory / 1024 / 1024
	
	return result, nil
}

// isAPMProcess 检查进程是否是APM相关进程
func (m *WindowsAPMMonitor) isAPMProcess(imageName, pidStr string) bool {
	// 直接匹配的APM工具
	apmTools := []string{"prometheus", "grafana", "jaeger", "zipkin", "elasticsearch", "kibana"}
	for _, tool := range apmTools {
		if strings.Contains(strings.ToLower(imageName), tool) {
			return true
		}
	}
	
	// 对于java.exe和node.exe，需要检查命令行参数
	if strings.Contains(strings.ToLower(imageName), "java") || strings.Contains(strings.ToLower(imageName), "node") {
		return m.checkProcessCommandLine(pidStr)
	}
	
	return false
}

// checkProcessCommandLine 检查进程命令行是否包含APM相关内容
func (m *WindowsAPMMonitor) checkProcessCommandLine(pidStr string) bool {
	cmd := exec.Command("wmic", "process", "where", fmt.Sprintf("ProcessId=%s", pidStr), "get", "CommandLine", "/format:value")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	cmdLine := strings.ToLower(string(output))
	apmKeywords := []string{
		"prometheus", "grafana", "jaeger", "zipkin", "elasticsearch", "kibana",
		"newrelic", "skywalking", "datadog", "appdynamics", "elastic-apm",
		"micrometer", "opentelemetry", "tracing", "metrics", "apm",
	}
	
	for _, keyword := range apmKeywords {
		if strings.Contains(cmdLine, keyword) {
			return true
		}
	}
	
	return false
}

// identifyAPMTool 识别APM工具类型
func (m *WindowsAPMMonitor) identifyAPMTool(imageName, pidStr string) string {
	imageLower := strings.ToLower(imageName)
	
	if strings.Contains(imageLower, "prometheus") {
		return "Prometheus"
	} else if strings.Contains(imageLower, "grafana") {
		return "Grafana"
	} else if strings.Contains(imageLower, "jaeger") {
		return "Jaeger"
	} else if strings.Contains(imageLower, "zipkin") {
		return "Zipkin"
	} else if strings.Contains(imageLower, "elasticsearch") {
		return "Elasticsearch"
	} else if strings.Contains(imageLower, "kibana") {
		return "Kibana"
	}
	
	// 对于java.exe和node.exe，通过命令行识别
	if strings.Contains(imageLower, "java") || strings.Contains(imageLower, "node") {
		cmd := exec.Command("wmic", "process", "where", fmt.Sprintf("ProcessId=%s", pidStr), "get", "CommandLine", "/format:value")
		output, err := cmd.Output()
		if err == nil {
			cmdLine := strings.ToLower(string(output))
			if strings.Contains(cmdLine, "prometheus") {
				return "Prometheus"
			} else if strings.Contains(cmdLine, "grafana") {
				return "Grafana"
			} else if strings.Contains(cmdLine, "jaeger") {
				return "Jaeger"
			} else if strings.Contains(cmdLine, "zipkin") {
				return "Zipkin"
			} else if strings.Contains(cmdLine, "elasticsearch") {
				return "Elasticsearch"
			} else if strings.Contains(cmdLine, "kibana") {
				return "Kibana"
			} else if strings.Contains(cmdLine, "newrelic") {
				return "New Relic"
			} else if strings.Contains(cmdLine, "skywalking") {
				return "SkyWalking"
			} else if strings.Contains(cmdLine, "datadog") {
				return "Datadog"
			} else if strings.Contains(cmdLine, "appdynamics") {
				return "AppDynamics"
			} else if strings.Contains(cmdLine, "elastic-apm") {
				return "Elastic APM"
			}
		}
	}
	
	return "Unknown"
}

// parseMemoryUsage 解析内存使用量字符串
func (m *WindowsAPMMonitor) parseMemoryUsage(memStr string) int64 {
	// 移除逗号和"K"后缀
	memStr = strings.ReplaceAll(memStr, ",", "")
	memStr = strings.ReplaceAll(memStr, " K", "")
	memStr = strings.TrimSpace(memStr)
	
	if mem, err := strconv.ParseInt(memStr, 10, 64); err == nil {
		return mem * 1024 // 转换为字节
	}
	return 0
}

// getDetailedProcessInfo 获取详细的进程信息
func (m *WindowsAPMMonitor) getDetailedProcessInfo(pid int) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 打开进程句柄
	handle, _, _ := openProcess.Call(
		uintptr(PROCESS_QUERY_INFORMATION|PROCESS_VM_READ),
		uintptr(0),
		uintptr(pid),
	)
	
	if handle == 0 {
		return result, fmt.Errorf("failed to open process %d", pid)
	}
	defer closeHandle.Call(handle)
	
	// 获取内存信息
	var memCounters PROCESS_MEMORY_COUNTERS
	memCounters.Cb = uint32(unsafe.Sizeof(memCounters))
	
	ret, _, _ := getProcessMemoryInfo.Call(
		handle,
		uintptr(unsafe.Pointer(&memCounters)),
		uintptr(memCounters.Cb),
	)
	
	if ret != 0 {
		result["working_set_size"] = int64(memCounters.WorkingSetSize)
		result["peak_working_set_size"] = int64(memCounters.PeakWorkingSetSize)
		result["pagefile_usage"] = int64(memCounters.PagefileUsage)
		result["peak_pagefile_usage"] = int64(memCounters.PeakPagefileUsage)
		result["page_fault_count"] = int64(memCounters.PageFaultCount)
	}
	
	// 获取进程时间信息
	var creationTime, exitTime, kernelTime, userTime FILETIME
	ret, _, _ = getProcessTimes.Call(
		handle,
		uintptr(unsafe.Pointer(&creationTime)),
		uintptr(unsafe.Pointer(&exitTime)),
		uintptr(unsafe.Pointer(&kernelTime)),
		uintptr(unsafe.Pointer(&userTime)),
	)
	
	if ret != 0 {
		// 转换FILETIME到时间
		creationTimeNs := (int64(creationTime.DwHighDateTime)<<32 + int64(creationTime.DwLowDateTime)) * 100
		creationTimeUnix := (creationTimeNs - 116444736000000000) / 10000000
		result["creation_time"] = time.Unix(creationTimeUnix, 0).Format(time.RFC3339)
		
		// 计算运行时间
		runningTime := time.Since(time.Unix(creationTimeUnix, 0))
		result["running_time_seconds"] = int64(runningTime.Seconds())
		result["running_time_formatted"] = runningTime.String()
		
		// CPU时间
		kernelTimeNs := (int64(kernelTime.DwHighDateTime)<<32 + int64(kernelTime.DwLowDateTime)) * 100
		userTimeNs := (int64(userTime.DwHighDateTime)<<32 + int64(userTime.DwLowDateTime)) * 100
		result["kernel_time_ns"] = kernelTimeNs
		result["user_time_ns"] = userTimeNs
		result["total_cpu_time_ns"] = kernelTimeNs + userTimeNs
	}
	
	return result, nil
}

// getPrometheusMetrics 获取Prometheus指标
func (m *WindowsAPMMonitor) getPrometheusMetrics() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 从配置获取Prometheus地址
	prometheusConfig, exists := m.agent.Config["prometheus"]
	if !exists {
		return nil, fmt.Errorf("prometheus config not found")
	}
	
	config := prometheusConfig.(map[string]interface{})
	host := config["host"].(string)
	port := int(config["port"].(float64))
	
	// 获取Prometheus状态
	statusURL := fmt.Sprintf("http://%s:%d/api/v1/status/config", host, port)
	if statusResp, err := m.sendHTTPRequest(statusURL, nil); err == nil {
		result["prometheus_status"] = statusResp
	} else {
		result["prometheus_error"] = err.Error()
	}
	
	// 获取目标信息
	targetsURL := fmt.Sprintf("http://%s:%d/api/v1/targets", host, port)
	if targetsResp, err := m.sendHTTPRequest(targetsURL, nil); err == nil {
		result["prometheus_targets"] = targetsResp
	}
	
	// 获取指标信息
	metricsURL := fmt.Sprintf("http://%s:%d/metrics", host, port)
	if metricsResp, err := m.sendHTTPRequest(metricsURL, nil); err == nil {
		result["prometheus_metrics"] = metricsResp
	}
	
	return result, nil
}

// getGrafanaMetrics 获取Grafana指标
func (m *WindowsAPMMonitor) getGrafanaMetrics() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 从配置获取Grafana地址
	grafanaConfig, exists := m.agent.Config["grafana"]
	if !exists {
		return nil, fmt.Errorf("grafana config not found")
	}
	
	config := grafanaConfig.(map[string]interface{})
	host := config["host"].(string)
	port := int(config["port"].(float64))
	
	// 准备认证头
	headers := make(map[string]string)
	if apiKey, exists := config["api_key"]; exists {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", apiKey.(string))
	}
	
	// 获取Grafana健康状态
	healthURL := fmt.Sprintf("http://%s:%d/api/health", host, port)
	if healthResp, err := m.sendHTTPRequest(healthURL, nil); err == nil {
		result["grafana_health"] = healthResp
	} else {
		result["grafana_error"] = err.Error()
	}
	
	// 获取数据源信息
	datasourcesURL := fmt.Sprintf("http://%s:%d/api/datasources", host, port)
	if dsResp, err := m.sendHTTPRequest(datasourcesURL, headers); err == nil {
		result["grafana_datasources"] = dsResp
	}
	
	// 获取仪表板信息
	dashboardsURL := fmt.Sprintf("http://%s:%d/api/search", host, port)
	if dashResp, err := m.sendHTTPRequest(dashboardsURL, headers); err == nil {
		result["grafana_dashboards"] = dashResp
	}
	
	return result, nil
}

// getJaegerMetrics 获取Jaeger指标
func (m *WindowsAPMMonitor) getJaegerMetrics() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 从配置获取Jaeger地址
	jaegerConfig, exists := m.agent.Config["jaeger"]
	if !exists {
		return nil, fmt.Errorf("jaeger config not found")
	}
	
	config := jaegerConfig.(map[string]interface{})
	host := config["host"].(string)
	port := int(config["port"].(float64))
	
	// 获取Jaeger服务列表
	servicesURL := fmt.Sprintf("http://%s:%d/api/services", host, port)
	if servicesResp, err := m.sendHTTPRequest(servicesURL, nil); err == nil {
		result["jaeger_services"] = servicesResp
	} else {
		result["jaeger_error"] = err.Error()
	}
	
	// 获取追踪信息
	tracesURL := fmt.Sprintf("http://%s:%d/api/traces?limit=10", host, port)
	if tracesResp, err := m.sendHTTPRequest(tracesURL, nil); err == nil {
		result["jaeger_traces"] = tracesResp
	}
	
	return result, nil
}

// getZipkinMetrics 获取Zipkin指标
func (m *WindowsAPMMonitor) getZipkinMetrics() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 从配置获取Zipkin地址
	zipkinConfig, exists := m.agent.Config["zipkin"]
	if !exists {
		return nil, fmt.Errorf("zipkin config not found")
	}
	
	config := zipkinConfig.(map[string]interface{})
	host := config["host"].(string)
	port := int(config["port"].(float64))
	
	// 获取Zipkin服务列表
	servicesURL := fmt.Sprintf("http://%s:%d/api/v2/services", host, port)
	if servicesResp, err := m.sendHTTPRequest(servicesURL, nil); err == nil {
		result["zipkin_services"] = servicesResp
	} else {
		result["zipkin_error"] = err.Error()
	}
	
	// 获取追踪信息
	tracesURL := fmt.Sprintf("http://%s:%d/api/v2/traces?limit=10", host, port)
	if tracesResp, err := m.sendHTTPRequest(tracesURL, nil); err == nil {
		result["zipkin_traces"] = tracesResp
	}
	
	return result, nil
}

// getElasticAPMMetrics 获取Elastic APM指标
func (m *WindowsAPMMonitor) getElasticAPMMetrics() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 从配置获取Elastic APM地址
	elasticConfig, exists := m.agent.Config["elastic_apm"]
	if !exists {
		return nil, fmt.Errorf("elastic apm config not found")
	}
	
	config := elasticConfig.(map[string]interface{})
	host := config["host"].(string)
	port := int(config["port"].(float64))
	
	// 准备认证头
	headers := make(map[string]string)
	if apiKey, exists := config["api_key"]; exists {
		headers["Authorization"] = fmt.Sprintf("ApiKey %s", apiKey.(string))
	}
	
	// 获取APM服务器信息
	infoURL := fmt.Sprintf("http://%s:%d/", host, port)
	if infoResp, err := m.sendHTTPRequest(infoURL, nil); err == nil {
		result["elastic_apm_info"] = infoResp
	} else {
		result["elastic_apm_error"] = err.Error()
	}
	
	return result, nil
}

// getNewRelicMetrics 获取New Relic指标
func (m *WindowsAPMMonitor) getNewRelicMetrics() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 从配置获取New Relic信息
	newRelicConfig, exists := m.agent.Config["new_relic"]
	if !exists {
		return nil, fmt.Errorf("new relic config not found")
	}
	
	config := newRelicConfig.(map[string]interface{})
	apiKey := config["api_key"].(string)
	
	// 准备认证头
	headers := map[string]string{
		"Api-Key": apiKey,
	}
	
	// 获取应用程序列表
	appsURL := "https://api.newrelic.com/v2/applications.json"
	if appsResp, err := m.sendHTTPRequest(appsURL, headers); err == nil {
		result["newrelic_applications"] = appsResp
	} else {
		result["newrelic_error"] = err.Error()
	}
	
	return result, nil
}

// getSkyWalkingMetrics 获取SkyWalking指标
func (m *WindowsAPMMonitor) getSkyWalkingMetrics() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// 从配置获取SkyWalking地址
	skywalkingConfig, exists := m.agent.Config["skywalking"]
	if !exists {
		return nil, fmt.Errorf("skywalking config not found")
	}
	
	config := skywalkingConfig.(map[string]interface{})
	host := config["host"].(string)
	port := int(config["port"].(float64))
	
	// 获取SkyWalking健康状态
	healthURL := fmt.Sprintf("http://%s:%d/graphql", host, port)
	
	// GraphQL查询获取服务列表
	query := `{"query":"query { getAllServices(duration: { start: \"2023-01-01 00:00:00\", end: \"2023-12-31 23:59:59\", step: DAY }) { key label } }"}`
	
	req, err := http.NewRequest("POST", healthURL, strings.NewReader(query))
	if err == nil {
		req.Header.Set("Content-Type", "application/json")
		resp, err := m.httpClient.Do(req)
		if err == nil {
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				var graphqlResp map[string]interface{}
				if json.Unmarshal(body, &graphqlResp) == nil {
					result["skywalking_services"] = graphqlResp
				}
			}
		} else {
			result["skywalking_error"] = err.Error()
		}
	}
	
	return result, nil
}

// 重写collectMetrics方法以使用Windows特定的实现
func (m *WindowsAPMMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})
	
	// 添加基本信息
	metrics["collection_time"] = time.Now().Format(time.RFC3339)
	
	// 使用Windows特定的方法收集指标
	if processInfo, err := m.getAPMProcessInfo(); err == nil {
		for k, v := range processInfo {
			metrics[k] = v
		}
	} else {
		m.agent.Logger.Error("Failed to get process info: %v", err)
		// 如果获取进程信息失败，设置基本状态
		metrics["apm_processes_running"] = false
		metrics["process_error"] = err.Error()
	}
	
	// 收集各种APM工具的指标
	if prometheusMetrics, err := m.getPrometheusMetrics(); err == nil {
		metrics["prometheus"] = prometheusMetrics
	} else {
		m.agent.Logger.Error("Failed to get Prometheus metrics: %v", err)
	}
	
	if grafanaMetrics, err := m.getGrafanaMetrics(); err == nil {
		metrics["grafana"] = grafanaMetrics
	} else {
		m.agent.Logger.Error("Failed to get Grafana metrics: %v", err)
	}
	
	if jaegerMetrics, err := m.getJaegerMetrics(); err == nil {
		metrics["jaeger"] = jaegerMetrics
	} else {
		m.agent.Logger.Error("Failed to get Jaeger metrics: %v", err)
	}
	
	if zipkinMetrics, err := m.getZipkinMetrics(); err == nil {
		metrics["zipkin"] = zipkinMetrics
	} else {
		m.agent.Logger.Error("Failed to get Zipkin metrics: %v", err)
	}
	
	if elasticMetrics, err := m.getElasticAPMMetrics(); err == nil {
		metrics["elastic_apm"] = elasticMetrics
	} else {
		m.agent.Logger.Error("Failed to get Elastic APM metrics: %v", err)
	}
	
	if newRelicMetrics, err := m.getNewRelicMetrics(); err == nil {
		metrics["new_relic"] = newRelicMetrics
	} else {
		m.agent.Logger.Error("Failed to get New Relic metrics: %v", err)
	}
	
	if skywalkingMetrics, err := m.getSkyWalkingMetrics(); err == nil {
		metrics["skywalking"] = skywalkingMetrics
	} else {
		m.agent.Logger.Error("Failed to get SkyWalking metrics: %v", err)
	}
	
	// 如果大部分指标收集失败，回退到模拟数据
	if len(metrics) <= 1 { // 只有基本信息
		m.agent.Logger.Warn("Most metric collection failed, falling back to simulated data")
		return m.APMMonitor.collectMetrics()
	}
	
	return metrics
}