package prober

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPProber HTTP 探针
type HTTPProber struct{}

// NewHTTPProber 创建 HTTP 探针
func NewHTTPProber() *HTTPProber {
	return &HTTPProber{}
}

// Type 返回探针类型
func (p *HTTPProber) Type() string {
	return "http"
}

// ConfigSchema 返回配置表单 schema
func (p *HTTPProber) ConfigSchema() map[string]FieldSchema {
	return map[string]FieldSchema{
		"url": {
			Type:        "string",
			Label:       "URL 地址",
			Required:    true,
			Placeholder: "http://example.com/api/health",
		},
		"method": {
			Type:         "select",
			Label:        "请求方法",
			Required:     false,
			DefaultValue: "GET",
			Options: []Option{
				{Value: "GET", Label: "GET"},
				{Value: "POST", Label: "POST"},
				{Value: "HEAD", Label: "HEAD"},
			},
		},
		"headers": {
			Type:        "string",
			Label:       "请求头",
			Required:    false,
			Placeholder: "Authorization: Bearer xxx",
			Hint:        "每行一个，格式: Key: Value",
		},
		"body": {
			Type:     "string",
			Label:    "请求体",
			Required: false,
			ShowWhen: map[string]any{"method": "POST"},
		},
		"expected_status": {
			Type:         "number",
			Label:        "期望状态码",
			Required:     false,
			DefaultValue: 200,
		},
		"expected_body": {
			Type:        "string",
			Label:       "期望响应包含",
			Required:    false,
			Placeholder: "ok",
			Hint:        "响应体需包含此字符串",
		},
		"insecure_skip_verify": {
			Type:         "boolean",
			Label:        "跳过证书验证",
			Required:     false,
			DefaultValue: false,
		},
	}
}

// Probe 执行探测
func (p *HTTPProber) Probe(ctx context.Context, target Target) (*ProbeResult, error) {
	start := time.Now()

	url := getStringConfig(target.Config, "url", "")
	if url == "" {
		return &ProbeResult{
			Success:   false,
			Latency:   time.Since(start),
			Message:   "URL 不能为空",
			CheckedAt: time.Now(),
		}, nil
	}

	method := getStringConfig(target.Config, "method", "GET")
	insecureSkipVerify := getBoolConfig(target.Config, "insecure_skip_verify", false)

	client := &http.Client{
		Timeout: target.Timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
		},
	}

	var bodyReader io.Reader
	if method == "POST" {
		body := getStringConfig(target.Config, "body", "")
		if body != "" {
			bodyReader = strings.NewReader(body)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return &ProbeResult{
			Success:   false,
			Latency:   time.Since(start),
			Message:   fmt.Sprintf("创建请求失败: %v", err),
			CheckedAt: time.Now(),
		}, nil
	}

	// 设置自定义 headers
	headers := getStringConfig(target.Config, "headers", "")
	if headers != "" {
		lines := strings.Split(headers, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}
	}

	resp, err := client.Do(req)
	latency := time.Since(start)

	if err != nil {
		return &ProbeResult{
			Success:   false,
			Latency:   latency,
			Message:   fmt.Sprintf("请求失败: %v", err),
			CheckedAt: time.Now(),
		}, nil
	}
	defer resp.Body.Close()

	// 读取响应体
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 10240)) // 最多读取 10KB
	bodyStr := string(bodyBytes)

	metrics := make(map[string]any)
	metrics["status_code"] = resp.StatusCode
	metrics["content_length"] = resp.ContentLength
	metrics["proto"] = resp.Proto

	var warnings []string

	// 检查状态码
	expectedStatus := getIntConfig(target.Config, "expected_status", 200)
	if resp.StatusCode != expectedStatus {
		return &ProbeResult{
			Success:   false,
			Latency:   latency,
			Message:   fmt.Sprintf("状态码 %d 不符合期望 %d", resp.StatusCode, expectedStatus),
			Metrics:   metrics,
			CheckedAt: time.Now(),
		}, nil
	}

	// 检查响应体
	expectedBody := getStringConfig(target.Config, "expected_body", "")
	if expectedBody != "" && !strings.Contains(bodyStr, expectedBody) {
		return &ProbeResult{
			Success:   false,
			Latency:   latency,
			Message:   fmt.Sprintf("响应体不包含期望内容: %s", expectedBody),
			Metrics:   metrics,
			CheckedAt: time.Now(),
		}, nil
	}

	message := fmt.Sprintf("HTTP %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))

	return &ProbeResult{
		Success:   true,
		Latency:   latency,
		Message:   message,
		Metrics:   metrics,
		CheckedAt: time.Now(),
		Warnings:  warnings,
	}, nil
}

// Validate 验证目标配置
func (p *HTTPProber) Validate(target Target) error {
	if _, ok := target.Config["url"]; !ok {
		return fmt.Errorf("缺少必填字段: url")
	}
	return nil
}
