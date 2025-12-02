package prober

import (
	"strings"
)

// getStringConfig 获取字符串配置
func getStringConfig(config map[string]any, key string, defaultValue string) string {
	if v, ok := config[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultValue
}

// getIntConfig 获取整数配置
func getIntConfig(config map[string]any, key string, defaultValue int) int {
	if v, ok := config[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		case float64:
			return int(val)
		}
	}
	return defaultValue
}

// getFloatConfig 获取浮点数配置
func getFloatConfig(config map[string]any, key string, defaultValue float64) float64 {
	if v, ok := config[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		case int64:
			return float64(val)
		}
	}
	return defaultValue
}

// getBoolConfig 获取布尔配置
func getBoolConfig(config map[string]any, key string, defaultValue bool) bool {
	if v, ok := config[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// getStringSliceConfig 获取字符串数组配置（逗号分隔）
func getStringSliceConfig(config map[string]any, key string) []string {
	if v, ok := config[key]; ok {
		if s, ok := v.(string); ok && s != "" {
			parts := strings.Split(s, ",")
			result := make([]string, 0, len(parts))
			for _, p := range parts {
				if trimmed := strings.TrimSpace(p); trimmed != "" {
					result = append(result, trimmed)
				}
			}
			return result
		}
	}
	return nil
}
