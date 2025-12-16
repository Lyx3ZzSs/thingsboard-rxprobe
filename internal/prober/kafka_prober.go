package prober

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

// KafkaProber Kafka 探针
type KafkaProber struct{}

// NewKafkaProber 创建 Kafka 探针
func NewKafkaProber() *KafkaProber {
	return &KafkaProber{}
}

// Type 返回探针类型
func (p *KafkaProber) Type() string {
	return "kafka"
}

// ConfigSchema 返回配置表单 schema
func (p *KafkaProber) ConfigSchema() map[string]FieldSchema {
	return map[string]FieldSchema{
		"brokers": {
			Type:        "string",
			Label:       "Broker 地址",
			Required:    true,
			Placeholder: "kafka1:9092,kafka2:9092,kafka3:9092",
			Hint:        "多个地址用逗号分隔",
		},
		"sasl_enabled": {
			Type:         "boolean",
			Label:        "启用 SASL 认证",
			Required:     false,
			DefaultValue: false,
		},
		"sasl_mechanism": {
			Type:         "select",
			Label:        "SASL 机制",
			Required:     false,
			DefaultValue: "PLAIN",
			ShowWhen:     map[string]any{"sasl_enabled": true},
			Options: []Option{
				{Value: "PLAIN", Label: "PLAIN"},
				{Value: "SCRAM-SHA-256", Label: "SCRAM-SHA-256"},
				{Value: "SCRAM-SHA-512", Label: "SCRAM-SHA-512"},
			},
		},
		"sasl_username": {
			Type:     "string",
			Label:    "SASL 用户名",
			Required: false,
			ShowWhen: map[string]any{"sasl_enabled": true},
		},
		"sasl_password": {
			Type:     "password",
			Label:    "SASL 密码",
			Required: false,
			ShowWhen: map[string]any{"sasl_enabled": true},
		},
		"tls_enabled": {
			Type:         "boolean",
			Label:        "启用 TLS",
			Required:     false,
			DefaultValue: false,
		},
	}
}

// Probe 执行探测
func (p *KafkaProber) Probe(ctx context.Context, target Target) (*ProbeResult, error) {
	start := time.Now()

	brokers := getStringSliceConfig(target.Config, "brokers")
	if len(brokers) == 0 {
		return &ProbeResult{
			Success:   false,
			Latency:   time.Since(start),
			Message:   "未配置 Broker 地址",
			CheckedAt: time.Now(),
		}, nil
	}

	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	config.Admin.Timeout = target.Timeout
	config.Net.DialTimeout = target.Timeout
	config.Metadata.Timeout = target.Timeout

	// SASL 配置
	if getBoolConfig(target.Config, "sasl_enabled", false) {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = getStringConfig(target.Config, "sasl_username", "")
		config.Net.SASL.Password = getStringConfig(target.Config, "sasl_password", "")

		mechanism := getStringConfig(target.Config, "sasl_mechanism", "PLAIN")
		switch mechanism {
		case "PLAIN":
			config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		case "SCRAM-SHA-256":
			config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
			config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
				return &XDGSCRAMClient{HashGeneratorFcn: SHA256}
			}
		case "SCRAM-SHA-512":
			config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
			config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
				return &XDGSCRAMClient{HashGeneratorFcn: SHA512}
			}
		}
	}

	// TLS 配置
	if getBoolConfig(target.Config, "tls_enabled", false) {
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	// 创建 Admin Client
	admin, err := sarama.NewClusterAdmin(brokers, config)
	if err != nil {
		return &ProbeResult{
			Success:   false,
			Latency:   time.Since(start),
			Message:   fmt.Sprintf("连接失败: %v", err),
			CheckedAt: time.Now(),
		}, nil
	}
	defer admin.Close()

	// 获取 Broker 列表验证集群可用性
	brokerList, _, err := admin.DescribeCluster()
	if err != nil {
		return &ProbeResult{
			Success:   false,
			Latency:   time.Since(start),
			Message:   fmt.Sprintf("获取集群信息失败: %v", err),
			CheckedAt: time.Now(),
		}, nil
	}

	return &ProbeResult{
		Success:   true,
		Latency:   time.Since(start),
		Message:   fmt.Sprintf("Kafka 集群服务可用，%d 个 Broker 在线", len(brokerList)),
		CheckedAt: time.Now(),
	}, nil
}

// Validate 验证目标配置
func (p *KafkaProber) Validate(target Target) error {
	if _, ok := target.Config["brokers"]; !ok {
		return fmt.Errorf("缺少必填字段: brokers")
	}
	return nil
}
