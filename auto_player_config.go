package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// 配置文件结构
type AutoPlayerConfigFile struct {
	Enabled          bool    `json:"enabled"`
	AutoDiscard      bool    `json:"autoDiscard"`
	AutoMeld         bool    `json:"autoMeld"`
	AutoRiichi       bool    `json:"autoRiichi"`
	AutoAgari        bool    `json:"autoAgari"`
	MinConfidence    float64 `json:"minConfidence"`
	DefenseThreshold float64 `json:"defenseThreshold"`
	DelaySeconds     float64 `json:"delaySeconds"`
	ConfirmActions   bool    `json:"confirmActions"`
	Strategy         string  `json:"strategy"`
}

const (
	configFileName = "auto_player_config.json"
)

// 加载配置文件
func LoadAutoPlayerConfig() error {
	configPath := getConfigPath()
	
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置文件不存在，使用默认配置
		return SaveDefaultConfig()
	}
	
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}
	
	var configFile AutoPlayerConfigFile
	if err := json.Unmarshal(data, &configFile); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}
	
	// 验证配置
	if err := validateConfig(configFile); err != nil {
		return fmt.Errorf("配置文件验证失败: %v", err)
	}
	
	// 应用配置
	config := AutoPlayerConfig{
		Enabled:          configFile.Enabled,
		AutoDiscard:      configFile.AutoDiscard,
		AutoMeld:         configFile.AutoMeld,
		AutoRiichi:       configFile.AutoRiichi,
		AutoAgari:        configFile.AutoAgari,
		MinConfidence:    configFile.MinConfidence,
		DefenseThreshold: configFile.DefenseThreshold,
		DelaySeconds:     configFile.DelaySeconds,
		ConfirmActions:   configFile.ConfirmActions,
		Strategy:         configFile.Strategy,
	}
	
	SetAutoPlayerConfig(config)
	return nil
}

// 保存配置文件
func SaveAutoPlayerConfig() error {
	config := GetAutoPlayerConfig()
	
	configFile := AutoPlayerConfigFile{
		Enabled:          config.Enabled,
		AutoDiscard:      config.AutoDiscard,
		AutoMeld:         config.AutoMeld,
		AutoRiichi:       config.AutoRiichi,
		AutoAgari:        config.AutoAgari,
		MinConfidence:    config.MinConfidence,
		DefenseThreshold: config.DefenseThreshold,
		DelaySeconds:     config.DelaySeconds,
		ConfirmActions:   config.ConfirmActions,
		Strategy:         config.Strategy,
	}
	
	data, err := json.MarshalIndent(configFile, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}
	
	configPath := getConfigPath()
	
	// 确保目录存在
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}
	
	if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}
	
	return nil
}

// 保存默认配置
func SaveDefaultConfig() error {
	defaultConfigFile := AutoPlayerConfigFile{
		Enabled:          false,
		AutoDiscard:      true,
		AutoMeld:         false,
		AutoRiichi:       false,
		AutoAgari:        true,
		MinConfidence:    0.7,
		DefenseThreshold: 0.15,
		DelaySeconds:     1.0,
		ConfirmActions:   true,
		Strategy:         "balanced",
	}
	
	data, err := json.MarshalIndent(defaultConfigFile, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化默认配置失败: %v", err)
	}
	
	configPath := getConfigPath()
	
	// 确保目录存在
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}
	
	if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入默认配置文件失败: %v", err)
	}
	
	// 应用默认配置
	config := AutoPlayerConfig{
		Enabled:          defaultConfigFile.Enabled,
		AutoDiscard:      defaultConfigFile.AutoDiscard,
		AutoMeld:         defaultConfigFile.AutoMeld,
		AutoRiichi:       defaultConfigFile.AutoRiichi,
		AutoAgari:        defaultConfigFile.AutoAgari,
		MinConfidence:    defaultConfigFile.MinConfidence,
		DefenseThreshold: defaultConfigFile.DefenseThreshold,
		DelaySeconds:     defaultConfigFile.DelaySeconds,
		ConfirmActions:   defaultConfigFile.ConfirmActions,
		Strategy:         defaultConfigFile.Strategy,
	}
	
	SetAutoPlayerConfig(config)
	return nil
}

// 验证配置
func validateConfig(config AutoPlayerConfigFile) error {
	if config.MinConfidence < 0.0 || config.MinConfidence > 1.0 {
		return fmt.Errorf("最小置信度必须在 0.0 到 1.0 之间")
	}
	
	if config.DefenseThreshold < 0.0 || config.DefenseThreshold > 1.0 {
		return fmt.Errorf("防守阈值必须在 0.0 到 1.0 之间")
	}
	
	if config.DelaySeconds < 0.0 || config.DelaySeconds > 10.0 {
		return fmt.Errorf("延迟时间必须在 0.0 到 10.0 秒之间")
	}
	
	validStrategies := []string{"aggressive", "balanced", "defensive"}
	valid := false
	for _, strategy := range validStrategies {
		if config.Strategy == strategy {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("策略必须是以下之一: %v", validStrategies)
	}
	
	return nil
}

// 获取配置文件路径
func getConfigPath() string {
	// 使用当前目录作为配置文件位置
	return filepath.Join(".", configFileName)
}

// 显示当前配置
func ShowAutoPlayerConfig() {
	config := GetAutoPlayerConfig()
	
	fmt.Println("🤖 自动出牌配置:")
	fmt.Printf("  启用状态: %t\n", config.Enabled)
	fmt.Printf("  自动切牌: %t\n", config.AutoDiscard)
	fmt.Printf("  自动鸣牌: %t\n", config.AutoMeld)
	fmt.Printf("  自动立直: %t\n", config.AutoRiichi)
	fmt.Printf("  自动和牌: %t\n", config.AutoAgari)
	fmt.Printf("  最小置信度: %.2f\n", config.MinConfidence)
	fmt.Printf("  防守阈值: %.2f\n", config.DefenseThreshold)
	fmt.Printf("  操作延迟: %.1f秒\n", config.DelaySeconds)
	fmt.Printf("  需要确认: %t\n", config.ConfirmActions)
	fmt.Printf("  策略类型: %s\n", config.Strategy)
}

// 重置为默认配置
func ResetAutoPlayerConfig() error {
	return SaveDefaultConfig()
}
