package main

import (
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// AuthConfig 权限验证配置
type AuthConfig struct {
	Enabled bool     `yaml:"enabled"`  // 是否启用权限验证
	APIKeys []string `yaml:"api_keys"` // API Key 列表
}

var (
	authConfig     *AuthConfig
	authConfigOnce sync.Once
)

// getAuthConfigPath 获取配置文件路径
// 优先使用环境变量 AUTH_CONFIG_PATH，否则使用默认值 auth.yaml
func getAuthConfigPath() string {
	if path := os.Getenv("AUTH_CONFIG_PATH"); path != "" {
		return path
	}
	return "auth.yaml"
}

// LoadAuthConfig 加载权限验证配置
// 如果配置文件不存在或加载失败，返回默认配置（不启用验证）
func LoadAuthConfig() *AuthConfig {
	authConfigOnce.Do(func() {
		authConfig = &AuthConfig{
			Enabled: false,
			APIKeys: []string{},
		}

		authConfigPath := getAuthConfigPath()

		// 检查配置文件是否存在
		if _, err := os.Stat(authConfigPath); os.IsNotExist(err) {
			logrus.Infof("权限验证配置文件 %s 不存在，默认禁用权限验证", authConfigPath)
			return
		}

		// 读取配置文件
		data, err := os.ReadFile(authConfigPath)
		if err != nil {
			logrus.Warnf("读取权限验证配置文件失败: %v，使用默认配置（禁用验证）", err)
			return
		}

		// 解析 YAML
		var config AuthConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			logrus.Warnf("解析权限验证配置文件失败: %v，使用默认配置（禁用验证）", err)
			return
		}

		authConfig = &config

		if authConfig.Enabled {
			logrus.Infof("权限验证已启用，API Keys 数量: %d", len(authConfig.APIKeys))
		} else {
			logrus.Info("权限验证配置已加载，但处于禁用状态")
		}
	})

	return authConfig
}

// IsAuthEnabled 检查是否启用了权限验证
func IsAuthEnabled() bool {
	config := LoadAuthConfig()
	return config.Enabled && len(config.APIKeys) > 0
}

// ValidateAPIKey 验证 API Key 是否有效
func ValidateAPIKey(apiKey string) bool {
	if !IsAuthEnabled() {
		return true // 未启用验证时，所有请求都通过
	}

	config := LoadAuthConfig()
	for _, key := range config.APIKeys {
		if key == apiKey {
			return true
		}
	}

	return false
}
