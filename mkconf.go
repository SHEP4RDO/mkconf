package mkconf

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	reader "mkconf/readers"
)

// ConfigSettings represents the configuration settings for a specific configuration file.
type ConfigSettings struct {
	configName     string                 // Name of the configuration
	configPath     string                 // Path to the configuration file
	configFullPath string                 // Full path to the configuration file
	configType     string                 // Type of the configuration file (e.g., JSON, YAML)
	Reader         reader.ConfigReader    // ConfigReader implementation for reading the configuration
	checkSec       int                    // Interval in seconds for checking configuration changes
	repeatSec      int                    // Interval in seconds for repeated configuration checks
	lastConfigHash string                 // Hash of the last known configuration file content
	configMAP      map[string]interface{} // Map representation of the configuration
	config         interface{}            // Instance of the configuration struct
	mu             sync.Mutex             // Mutex for synchronizing access to configuration data
	ctx            context.Context        // Context for cancellation of configuration monitoring
	cancel         context.CancelFunc     // Cancel function to stop configuration monitoring
	waitGroup      *sync.WaitGroup        // WaitGroup to wait for the completion of monitoring goroutines

	enableChangeValidation bool // Flag to enable change validation for the configuration
	enableChangeTracking   bool // Flag to enable change tracking for the configuration

	ch_ChangeValidation chan struct{} // Channel for signaling change validation
	Ch_ConfigChanged    chan string   // Channel for signaling configuration changes
	Ch_ConfigTracking   chan string   // Channel for signaling configuration tracking
}

// ConfigList represents a collection of configuration settings.
type ConfigList struct {
	settingsMutex sync.Mutex                   // Mutex for synchronizing access to the settings map
	settings      map[string]*ConfigSettings   // Map of configuration settings with configName as the key
	changeLogs    map[string][]ConfigChangeLog // Map of configuration change logs with configName as the key
	logMutex      sync.Mutex                   // Mutex for synchronizing access to the changeLogs map
}

// NewConfigList creates a new ConfigList instance.
func NewConfigList() *ConfigList {
	list := &ConfigList{}
	list.settings = make(map[string]*ConfigSettings)
	return list
}

// GetSettings returns the ConfigSettings for the specified configuration file name.
func (c *ConfigList) GetSettings(fileName string) *ConfigSettings {
	return c.settings[fileName]
}

// SetConfigName sets the name of the configuration.
func (c *ConfigSettings) SetConfigName(fileName string) *ConfigSettings {
	c.configName = fileName
	return c
}

// SetConfigPath sets the path of the configuration file.
func (c *ConfigSettings) SetConfigPath(path string) *ConfigSettings {
	c.configPath = path
	return c
}

// SetChangeValidation sets the flag to enable or disable change validation for the configuration.
func (c *ConfigSettings) SetChangeValidation(isValid bool) *ConfigSettings {
	c.enableChangeValidation = isValid
	return c
}

// SetConfigType sets the type of the configuration file (e.g., JSON, YAML).
func (c *ConfigSettings) SetConfigType(fileType string) *ConfigSettings {
	c.configType = fileType
	return c
}

// SetConfigFullpath sets the full path of the configuration file.
func (c *ConfigSettings) SetConfigFullpath(fullPath string) *ConfigSettings {
	c.configFullPath = fullPath
	return c
}

// GetChangesChan returns the channel for signaling configuration changes for the specified configuration name.
func (c *ConfigList) GetChangesChan(configName string) chan string {
	return c.settings[configName].Ch_ConfigChanged
}

// SetReader sets the ConfigReader for reading the configuration.
func (c *ConfigSettings) SetReader(reader reader.ConfigReader) *ConfigSettings {
	c.Reader = reader
	return c
}

// SetCheckSec sets the repeat interval in seconds for checking configuration changes.
func (c *ConfigSettings) SetCheckSec(repeatInterval int) *ConfigSettings {
	c.checkSec = repeatInterval
	return c
}

// SetRepeatSec sets the check interval in seconds for repeated configuration checks.
func (c *ConfigSettings) SetRepeatSec(checkInterval int) *ConfigSettings {
	c.repeatSec = checkInterval
	return c
}

// setHash sets the last known hash of the configuration file.
func (c *ConfigSettings) SetHash(hash string) *ConfigSettings {
	c.lastConfigHash = hash
	return c
}

// SetChangeTracking sets the flag to enable or disable change tracking for the configuration.
func (c *ConfigSettings) SetChangeTracking(mode bool) *ConfigSettings {
	c.enableChangeTracking = mode
	return c
}

// LoadConfig loads the configuration with the specified name and populates the provided interface.
// It automatically selects the appropriate reader based on the file type if the reader is not set.
// It returns an error if the configuration cannot be loaded or if there is an issue with the reader.
func (c *ConfigList) LoadConfig(configName string, v interface{}) error {
	if c.settings[configName].Reader == nil {
		reader := c.settings[configName].checkReader()
		if reader == nil {
			return fmt.Errorf("%v error while setting reader type - check your config file type", configName)
		}

		c.settings[configName].SetReader(reader)
	}
	err := c.settings[configName].Reader.ReadConfig(c.settings[configName].configFullPath, v)
	if err != nil {
		return fmt.Errorf("load config %v: error while read config: %v", configName, err)
	}
	c.settings[configName].config = v
	return nil
}

// UpdateConfig updates the configuration with the specified name by applying changes from the provided interface.
// It first stops the change monitoring, performs the update, and then restarts the change monitoring.
// It returns an error if the update fails or if the reader is not set for the configuration.
func (c *ConfigList) UpdateConfig(configName string, v interface{}) error {
	c.settingsMutex.Lock()
	defer c.settingsMutex.Unlock()

	settings, ok := c.settings[configName]
	if !ok {
		return fmt.Errorf("config with name %s not found", configName)
	}

	if settings.Reader == nil {
		return fmt.Errorf("reader not set for config %s", configName)
	}

	c.StopChangeMonitoring(configName)
	defer c.StartChangeMonitoring(configName, v)

	err := settings.Reader.UpdateConfig(settings.configFullPath, v)
	if err != nil {
		return fmt.Errorf("update config %s: %v", configName, err)
	}

	err = c.LoadConfig(configName, settings.config)
	if err != nil {
		return fmt.Errorf("reload config %s: %v", configName, err)
	}

	return nil
}

// checkReader selects a ConfigReader based on the file type and returns it.
// It is used to automatically set the reader if it is not explicitly provided.
func (s *ConfigSettings) checkReader() reader.ConfigReader {
	_type := strings.ToLower(s.configType)
	switch _type {
	case ".json", ".mk.json":
		return &reader.JSONConfigReader{}
	case ".xml", ".mk.xml":
		return &reader.XMLConfigReader{}
	case ".yaml", ".yml", ".mk.yaml", ".mk.yml":
		return &reader.YAMLConfigReader{}
	case ".toml", ".mk.toml":
		return &reader.TOMLConfigReader{}
	case ".ini", ".mk.ini":
		return &reader.INIConfigReader{}
	default:
		return nil
	}
}

// AddConfigList adds a new configuration to the ConfigList with the provided name, path, type, and interface.
// It initializes the configuration settings, including channels and readers, and calculates the initial hash.
// Returns an error if there's an issue adding the new configuration.
func (c *ConfigList) AddConfigList(configName, configPath, configType string, v interface{}) error {
	var err error
	settings := ConfigSettings{
		configName:             configName,
		configPath:             configPath,
		configType:             configType,
		enableChangeValidation: false,
		enableChangeTracking:   false,
		checkSec:               1,
		repeatSec:              10,
		ch_ChangeValidation:    make(chan struct{}),
		Ch_ConfigChanged:       make(chan string),
		Ch_ConfigTracking:      make(chan string),
		waitGroup:              new(sync.WaitGroup),
	}
	c.changeLogs = map[string][]ConfigChangeLog{}
	c.settings[configName] = &settings
	fullConfigName := configName + configType
	fullPath := filepath.Join(configPath, fullConfigName)
	c.settings[configName].SetConfigPath(configPath).SetConfigFullpath(fullPath).defineReader()
	err = c.settings[configName].defineHash(v)
	if err != nil {
		return fmt.Errorf("mkconf: error add new config %v: %v", configName, err)
	}
	return nil
}

// defineHash calculates the hash of the configuration file and initializes the configuration map.
// It returns an error if there's an issue calculating the hash or converting the configuration to a map.
func (c *ConfigSettings) defineHash(v interface{}) error {
	var err error
	c.lastConfigHash, err = c.calculateFileHash(c.configFullPath)
	if err != nil {
		return fmt.Errorf("error calculate hash: %v", err)
	}
	configMap, _ := c.convertToMap(c.configFullPath)
	c.config = &v
	c.configMAP = configMap
	return nil
}

// defineReader sets the ConfigReader in ConfigSettings based on the configuration file type.
// It returns the updated ConfigSettings instance.
func (c *ConfigSettings) defineReader() *ConfigSettings {
	c.Reader = c.checkReader()
	return c
}

// convertToMap converts the configuration file to a map based on its type using the appropriate reader.
// It returns the map representation of the configuration file and an error if there's an issue.
func (c *ConfigSettings) convertToMap(fullPath string) (map[string]interface{}, error) {
	tmp := make(map[string]interface{})
	var err error

	switch reader := c.Reader.(type) {
	case *reader.JSONConfigReader:
		tmp, err = reader.ReadConfigToMap(fullPath)
	case *reader.XMLConfigReader:
		tmp, err = reader.ReadConfigToMap(fullPath)
	case *reader.YAMLConfigReader:
		tmp, err = reader.ReadConfigToMap(fullPath)
	case *reader.TOMLConfigReader:
		tmp, err = reader.ReadConfigToMap(fullPath)
	case *reader.INIConfigReader:
		tmp, err = reader.ReadConfigToMap(fullPath)
	default:
		return nil, fmt.Errorf("unsupported ConfigReader type - %v", reader)
	}

	if err != nil {
		return nil, fmt.Errorf("error converting config to map: %v", err)
	}

	return tmp, nil
}
