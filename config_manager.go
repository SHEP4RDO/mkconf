package mkconf

import (
	"fmt"
	"path/filepath"
	"strings"
)

type ConfigManager struct {
	configList *ConfigList
	configs    map[string]interface{}
}

func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		configList: NewConfigList(),
		configs:    make(map[string]interface{}),
	}
}

func (cm *ConfigManager) AddConfig(configName, configPath, configType string, configInterface interface{}) error {
	if _, ok := cm.configs[configName]; ok {
		return fmt.Errorf("config with name %s already exists", configName)
	}

	err := cm.configList.AddConfigList(configName, configPath, configType, configInterface)
	if err != nil {
		return err
	}

	cm.configs[configName] = configInterface
	return nil
}

func (cm *ConfigManager) GetSettings(configName string) *ConfigSettings {
	return cm.configList.settings[configName]
}
func (cm *ConfigManager) GetConfigList(configName string) *ConfigList {
	return cm.configList
}
func (cm *ConfigManager) GetConfig(configName string) (interface{}, error) {
	configInterface, ok := cm.configs[configName]
	if !ok {
		return nil, fmt.Errorf("config with name %s not found", configName)
	}
	return configInterface, nil
}
func (cm *ConfigManager) LoadConfigs() error {
	for configName, configInterface := range cm.configs {
		err := cm.configList.LoadConfig(configName, configInterface)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cm *ConfigManager) PrintConfigs() {
	for configName, configInterface := range cm.configs {
		fmt.Printf("%s - %v\n", configName, configInterface)
	}
}

func (cm *ConfigManager) LoadConfigsFromPath(configPath string, configNames []string, configInterfaces []interface{}) error {
	if len(configNames) != len(configInterfaces) {
		return fmt.Errorf("number of config names does not match number of config interfaces")
	}

	for i, configName := range configNames {
		configBase := strings.TrimSuffix(configName, filepath.Ext(configName))
		configType := filepath.Ext(configName)
		if configType == "" {
			return fmt.Errorf("unable to determine config type for %s", configName)
		}

		configType = strings.ToLower(configType)
		configInterface := configInterfaces[i]

		err := cm.AddConfig(configBase, configPath, configType, configInterface)
		if err != nil {
			return err
		}
	}

	return cm.LoadConfigs()
}

func (cm *ConfigManager) StartChangeMonitoring(configName string, v interface{}) error {
	return cm.configList.StartChangeMonitoring(configName, v)
}

func (cm *ConfigManager) StopChangeMonitoring(configName string) {
	cm.configList.StopChangeMonitoring(configName)
}

func (cm *ConfigManager) StartAllChangeMonitoring() {
	for configName, settings := range cm.configList.settings {
		if settings.enableChangeValidation {
			cm.StartChangeMonitoring(configName, settings.config)
		}
	}
}

func (cm *ConfigManager) StopAllChangeMonitoring() {
	for configName, settings := range cm.configList.settings {
		if !settings.enableChangeValidation {
			cm.StopChangeMonitoring(configName)
		}
	}
}

func (cm *ConfigManager) WatchForChanges() {
	for _, configName := range cm.configList.GetConfigNames() {
		settings := cm.configList.GetSettings(configName)
		if settings.enableChangeValidation {
			go func(ch <-chan string) {
				for name := range ch {
					fmt.Printf("Config '%v' has changed.\n", name)
				}
			}(settings.Ch_ConfigChanged)
		}
	}
}

func (c *ConfigList) GetConfigNames() []string {
	var names []string
	for name := range c.settings {
		names = append(names, name)
	}
	return names
}
func (cm *ConfigManager) UpdateConfig(configName string, configInterface interface{}) error {
	cm.configList.UpdateConfig(configName, configInterface)
	return nil
}

func (cm *ConfigManager) UpdateConfigs(configNames []string, configInterfaces []interface{}) error {
	if len(configNames) != len(configInterfaces) {
		return fmt.Errorf("number of config names does not match number of config interfaces")
	}

	for i, configName := range configNames {
		configInterface := configInterfaces[i]
		err := cm.UpdateConfig(configName, configInterface)
		if err != nil {
			return err
		}
	}

	return nil
}
