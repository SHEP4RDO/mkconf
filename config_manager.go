package mkconf

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ConfigManager is a manager that handles the configuration settings and interfaces for multiple configurations.
type ConfigManager struct {
	configList *ConfigList            // ConfigList instance to manage configuration settings and updates.
	configs    map[string]interface{} // Map to store configuration interfaces with their respective names.
}

// NewConfigManager creates a new instance of ConfigManager with an initialized ConfigList and an empty configs map.
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		configList: NewConfigList(),
		configs:    make(map[string]interface{}),
	}
}

// AddConfig adds a new configuration to the manager with the specified name, path, type, and interface.
// It associates the provided interface with the given name and sets up the corresponding configuration in the ConfigList.
// Returns an error if a configuration with the same name already exists.
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

// GetSettings returns the ConfigSettings associated with the specified configuration name.
func (cm *ConfigManager) GetSettings(configName string) *ConfigSettings {
	return cm.configList.settings[configName]
}

// GetConfigList returns the ConfigList instance associated with the ConfigManager.
func (cm *ConfigManager) GetConfigList(configName string) *ConfigList {
	return cm.configList
}

// GetConfig returns the configuration interface associated with the specified name.
// Returns an error if the configuration is not found.
func (cm *ConfigManager) GetConfig(configName string) (interface{}, error) {
	configInterface, ok := cm.configs[configName]
	if !ok {
		return nil, fmt.Errorf("config with name %s not found", configName)
	}
	return configInterface, nil
}

// LoadConfigs loads configurations for all registered interfaces in the manager.
// It iterates through each configuration and loads the corresponding settings using ConfigList.
// If any configuration fails to load, it logs an error and continues with the remaining configurations.
// Returns a slice of errors encountered during the loading process.
func (cm *ConfigManager) LoadConfigs() []error {
	var loadErrors []error

	for configName, configInterface := range cm.configs {
		err := cm.configList.LoadConfig(configName, configInterface)
		if err != nil {
			loadErrors = append(loadErrors, fmt.Errorf("error loading config %s: %v", configName, err))
		}
	}

	return loadErrors
}

// PrintConfigs prints the names and interface values of all registered configurations.
// Useful for debugging and checking the current state of registered configurations.
func (cm *ConfigManager) PrintConfigs() {
	for configName, configInterface := range cm.configs {
		fmt.Printf("%s - %v\n", configName, configInterface)
	}
}

// LoadConfigsFromPath loads configurations for specified names and interfaces from the given path.
// It adds configurations using AddConfig method and then loads them using LoadConfigs method.
// Returns a slice of errors encountered during the loading process, or nil if there are no errors.
func (cm *ConfigManager) LoadConfigsFromPath(configPath string, configNames []string, configInterfaces []interface{}) []error {
	if len(configNames) != len(configInterfaces) {
		return []error{fmt.Errorf("number of config names does not match number of config interfaces")}
	}

	var loadErrors []error

	for i, configName := range configNames {
		configBase := strings.TrimSuffix(configName, filepath.Ext(configName))
		configType := filepath.Ext(configName)
		if configType == "" {
			loadErrors = append(loadErrors, fmt.Errorf("unable to determine config type for %s", configName))
			continue
		}

		configType = strings.ToLower(configType)
		configInterface := configInterfaces[i]

		err := cm.AddConfig(configBase, configPath, configType, configInterface)
		if err != nil {
			loadErrors = append(loadErrors, err)
		}
	}

	if len(loadErrors) == 0 {
		return nil
	}

	return loadErrors
}

// StartChangeMonitoring starts change monitoring for a specific configuration.
// It enables the validation of changes and starts a goroutine to watch for changes in the configuration.
// The method returns an error if the specified configuration is not found.
func (cm *ConfigManager) StartChangeMonitoring(configName string, v interface{}) error {
	return cm.configList.StartChangeMonitoring(configName, v)
}

// StopChangeMonitoring stops change monitoring for a specific configuration.
// It cancels the change monitoring goroutine and waits for it to finish.
func (cm *ConfigManager) StopChangeMonitoring(configName string) {
	cm.configList.StopChangeMonitoring(configName)
}

// StartAllChangeMonitoring starts change monitoring for all configurations that have change validation enabled.
// It iterates through all configurations and starts change monitoring for each one.
func (cm *ConfigManager) StartAllChangeMonitoring() {
	for configName, settings := range cm.configList.settings {
		if settings.enableChangeValidation {
			cm.StartChangeMonitoring(configName, settings.config)
		}
	}
}

// StopAllChangeMonitoring stops change monitoring for all configurations that have change validation disabled.
// It iterates through all configurations and stops change monitoring for each one.
func (cm *ConfigManager) StopAllChangeMonitoring() {
	for configName, settings := range cm.configList.settings {
		if !settings.enableChangeValidation {
			cm.StopChangeMonitoring(configName)
		}
	}
}

// WatchForChanges starts watching for changes in configurations.
// It creates a goroutine for each configuration with change validation enabled.
// When a change is detected, it prints a message to the console.
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

// GetConfigNames returns a slice containing the names of all configurations in the ConfigList.
// It iterates through the settings map and collects the names of each configuration.
func (c *ConfigList) GetConfigNames() []string {
	var names []string
	for name := range c.settings {
		names = append(names, name)
	}
	return names
}

// UpdateConfig updates the specified configuration with a new interface.
// It delegates the update operation to the ConfigList.
func (cm *ConfigManager) UpdateConfig(configName string, configInterface interface{}) error {
	cm.configList.UpdateConfig(configName, configInterface)
	return nil
}

// UpdateConfigs updates multiple configurations with new interfaces.
// It iterates through the provided names and interfaces, updating each configuration.
// Returns an error if the number of names does not match the number of interfaces.
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
