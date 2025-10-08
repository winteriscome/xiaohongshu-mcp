package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
)

// AuthManager 权限管理器
type AuthManager struct {
	config     *AuthConfig
	configPath string
	mutex      sync.RWMutex
}

// NewAuthManager 创建权限管理器
func NewAuthManager(configPath string) *AuthManager {
	am := &AuthManager{
		configPath: configPath,
		config:     getDefaultAuthConfig(),
	}

	// 尝试加载配置文件
	if err := am.LoadConfig(); err != nil {
		logrus.Warnf("无法加载权限配置文件，使用默认配置: %v", err)
	}

	return am
}

// getDefaultAuthConfig 获取默认权限配置
func getDefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		Enabled: false, // 默认不启用权限验证
		APIKeys: []string{
			"default-key", // 默认API Key
		},
	}
}

// LoadConfig 加载配置文件
func (am *AuthManager) LoadConfig() error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	// 如果配置文件不存在，创建默认配置文件
	if _, err := os.Stat(am.configPath); os.IsNotExist(err) {
		return am.SaveConfig()
	}

	data, err := ioutil.ReadFile(am.configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config AuthConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	am.config = &config
	logrus.Infof("权限配置已加载: 启用状态=%v, API Key数量=%d", config.Enabled, len(config.APIKeys))
	return nil
}

// SaveConfig 保存配置文件
func (am *AuthManager) SaveConfig() error {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	// 确保配置目录存在
	dir := filepath.Dir(am.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	data, err := json.MarshalIndent(am.config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := ioutil.WriteFile(am.configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	logrus.Infof("权限配置已保存到: %s", am.configPath)
	return nil
}

// IsEnabled 检查权限验证是否启用
func (am *AuthManager) IsEnabled() bool {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	return am.config.Enabled
}

// ValidateAPIKey 验证API Key
func (am *AuthManager) ValidateAPIKey(apiKey string) bool {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	if !am.config.Enabled {
		return true // 如果未启用权限验证，默认允许所有操作
	}

	for _, validKey := range am.config.APIKeys {
		if validKey == apiKey {
			return true
		}
	}
	return false
}

// GetConfig 获取当前配置（只读）
func (am *AuthManager) GetConfig() AuthConfig {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	return *am.config
}
