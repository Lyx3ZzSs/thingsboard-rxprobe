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
		"consumer_groups": {
			Type:        "string",
			Label:       "监控消费组",
			Required:    false,
			Placeholder: "tb-core,tb-rule-engine",
			Hint:        "多个消费组用逗号分隔，监控其消费延迟",
		},
		"lag_threshold": {
			Type:         "number",
			Label:        "消费延迟告警阈值",
			Required:     false,
			DefaultValue: 10000,
			Hint:         "消息积压数量超过此值告警",
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

	metrics := make(map[string]any)
	var warnings []string

	// 1. 获取 Broker 列表
	brokerList, _, err := admin.DescribeCluster()
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("获取集群信息失败: %v", err))
	} else {
		brokerInfos := []map[string]any{}
		for _, broker := range brokerList {
			brokerInfos = append(brokerInfos, map[string]any{
				"id":   broker.ID(),
				"addr": broker.Addr(),
			})
		}
		metrics["brokers"] = brokerInfos
		metrics["broker_count"] = len(brokerList)
	}

	// 2. 获取 Topic 列表
	topics, err := admin.ListTopics()
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("获取 Topic 列表失败: %v", err))
	} else {
		topicCount := 0
		topicInfos := []map[string]any{}
		for name, detail := range topics {
			topicCount++
			topicInfos = append(topicInfos, map[string]any{
				"name":               name,
				"partitions":         detail.NumPartitions,
				"replication_factor": detail.ReplicationFactor,
			})
		}
		metrics["topic_count"] = topicCount
		// 只返回前20个topic避免数据过多
		if len(topicInfos) > 20 {
			topicInfos = topicInfos[:20]
		}
		metrics["topics"] = topicInfos
	}

	// 3. 检查消费组延迟
	consumerGroups := getStringSliceConfig(target.Config, "consumer_groups")
	if len(consumerGroups) > 0 {
		lagThreshold := int64(getIntConfig(target.Config, "lag_threshold", 10000))

		groupLags := make(map[string]int64)
		for _, group := range consumerGroups {
			offsets, err := admin.ListConsumerGroupOffsets(group, nil)
			if err != nil {
				continue
			}

			totalLag := int64(0)
			for topicName, partitions := range offsets.Blocks {
				for partition := range partitions {
					// 获取最新 offset
					newestOffset, err := getNewestOffset(brokers, config, topicName, partition)
					if err != nil {
						continue
					}

					consumerOffset := offsets.Blocks[topicName][partition].Offset
					if consumerOffset >= 0 && newestOffset > consumerOffset {
						totalLag += newestOffset - consumerOffset
					}
				}
			}

			groupLags[group] = totalLag
			if totalLag > lagThreshold {
				warnings = append(warnings, fmt.Sprintf("消费组 %s 延迟 %d 超过阈值", group, totalLag))
			}
		}
		metrics["consumer_group_lags"] = groupLags
	}

	// 4. 检查 under-replicated partitions
	if topics != nil {
		topicNames := make([]string, 0, len(topics))
		for name := range topics {
			topicNames = append(topicNames, name)
		}

		describeTopics, err := admin.DescribeTopics(topicNames)
		if err == nil {
			underReplicated := 0
			for _, topic := range describeTopics {
				for _, partition := range topic.Partitions {
					if len(partition.Isr) < len(partition.Replicas) {
						underReplicated++
					}
				}
			}
			metrics["under_replicated_partitions"] = underReplicated
			if underReplicated > 0 {
				warnings = append(warnings, fmt.Sprintf("存在 %d 个 under-replicated 分区", underReplicated))
			}
		}
	}

	latency := time.Since(start)
	brokerCount := 0
	if bc, ok := metrics["broker_count"].(int); ok {
		brokerCount = bc
	}
	message := fmt.Sprintf("Kafka 集群正常，%d 个 Broker 在线", brokerCount)

	if len(warnings) > 0 {
		message = fmt.Sprintf("存在告警: %v", warnings)
	}

	return &ProbeResult{
		Success:   true,
		Latency:   latency,
		Message:   message,
		Metrics:   metrics,
		CheckedAt: time.Now(),
		Warnings:  warnings,
	}, nil
}

// getNewestOffset 获取分区最新 offset
func getNewestOffset(brokers []string, config *sarama.Config, topic string, partition int32) (int64, error) {
	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		return 0, err
	}
	defer client.Close()

	return client.GetOffset(topic, partition, sarama.OffsetNewest)
}

// Validate 验证目标配置
func (p *KafkaProber) Validate(target Target) error {
	if _, ok := target.Config["brokers"]; !ok {
		return fmt.Errorf("缺少必填字段: brokers")
	}
	return nil
}
