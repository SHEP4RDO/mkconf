package mkconf

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
)

type ConfigSettings struct {
	configName             string
	configPath             string
	configFullPath         string
	configType             string
	enableChangeValidation bool
	enableChangeTracking   bool
	checkSec               int
	repeatSec              int
	Reader                 ConfigReader
	lastConfigHash         string
	mu                     sync.Mutex
	ctx                    context.Context
	cancel                 context.CancelFunc
	waitGroup              *sync.WaitGroup
	configMAP              map[string]interface{}
	config                 interface{}
	ch_ChangeValidation    chan struct{}
	Ch_ConfigChanged       chan struct{}
	Ch_ConfigTracking      chan struct{}
}

type ConfigManager struct {
	settings   map[string]*ConfigSettings
	changeLogs map[string][]ConfigChangeLog
	logMutex   sync.Mutex
}

func NewConfigManager() *ConfigManager {
	manager := &ConfigManager{}
	manager.settings = make(map[string]*ConfigSettings)
	return manager
}

func (c *ConfigManager) GetSettings(fileName string) *ConfigSettings {
	return c.settings[fileName]
}

func (c *ConfigSettings) SetConfigName(fileName string) *ConfigSettings {
	c.configName = fileName
	return c
}
func (c *ConfigSettings) SetConfigPath(path string) *ConfigSettings {
	c.configPath = path
	return c
}
func (c *ConfigSettings) SetChangeValidation(isValid bool) *ConfigSettings {
	c.enableChangeValidation = isValid
	return c
}
func (c *ConfigSettings) SetConfigType(fileType string) *ConfigSettings {
	c.configType = fileType
	return c
}
func (c *ConfigSettings) SetConfigFullpath(fullPath string) *ConfigSettings {
	c.configFullPath = fullPath
	return c
}
func (c *ConfigManager) GetChangesChan(configName string) chan struct{} {
	return c.settings[configName].Ch_ConfigChanged
}

func (c *ConfigSettings) SetReader(reader ConfigReader) *ConfigSettings {
	c.Reader = reader
	return c
}

func (c *ConfigSettings) SetCheckSec(repeat_interval int) *ConfigSettings {
	c.checkSec = repeat_interval
	return c
}
func (c *ConfigSettings) SetRepeatSec(check_interval int) *ConfigSettings {
	c.repeatSec = check_interval
	return c
}
func (c *ConfigSettings) setHash(hash string) *ConfigSettings {
	c.lastConfigHash = hash
	return c
}
func (c *ConfigSettings) SetChangeTracking(mode bool) *ConfigSettings {
	c.enableChangeTracking = mode
	return c
}
func (c *ConfigManager) LoadConfig(configName string, v interface{}) error {
	if c.settings[configName].Reader == nil {
		reader := c.settings[configName].checkReader()
		if reader == nil {
			return fmt.Errorf("error while setting reader type - check your config file type")
		}

		c.settings[configName].SetReader(reader)
	}
	err := c.settings[configName].Reader.ReadConfig(c.settings[configName].configFullPath, &v)
	if err != nil {
		return fmt.Errorf("load config: error while read config: %v\n", err)
	}
	if c.settings[configName].enableChangeValidation {
		hash, err := c.settings[configName].calculateFileHash(c.settings[configName].configFullPath)
		if err != nil {
			return fmt.Errorf("load config: error while calculate hash: %v", err)
		}
		c.settings[configName].setHash(hash)
		c.StartChangeMonitoring(configName, &v)
	}
	c.settings[configName].config = v
	return nil
}
func (s *ConfigSettings) checkReader() ConfigReader {
	switch s.configType {
	case ".json", ".JSON":
		return &JSONConfigReader{}
	case ".xml", ".XML":
		return &XMLConfigReader{}
	case ".yaml", ".yml", ".YAML", ".YML":
		return &YAMLConfigReader{}
	case ".toml", ".TOML":
		return &TOMLConfigReader{}
	case ".ini", ".INI":
		return &INIConfigReader{}
	default:
		return nil
	}
}

func (c *ConfigManager) AddConfig(configName, configPath, configType string, v interface{}) error {
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
		Ch_ConfigChanged:       make(chan struct{}),
		Ch_ConfigTracking:      make(chan struct{}),
		waitGroup:              new(sync.WaitGroup),
	}
	c.changeLogs = map[string][]ConfigChangeLog{}
	c.settings[configName] = &settings
	fullConfigName := configName + configType
	fullPath := filepath.Join(configPath, fullConfigName)
	c.settings[configName].SetConfigPath(configPath).SetConfigFullpath(fullPath).defineReader()
	err = c.settings[configName].defineHash(configName, v)
	if err != nil {
		return fmt.Errorf("mkconf: error add new config: %v", err)
	}
	return nil
}

func (c *ConfigSettings) defineHash(configName string, v interface{}) error {
	var err error
	c.lastConfigHash, err = c.calculateFileHash(c.configFullPath)
	if err != nil {
		return fmt.Errorf("error calculate hash: %v", err)
	}
	configMap, err := c.convertToMap(c.configFullPath)
	if err != nil {
		return fmt.Errorf("error convert map: %v", err)
	}
	c.config = v
	c.configMAP = configMap
	return nil
}

func (c *ConfigSettings) defineReader() *ConfigSettings {
	c.Reader = c.checkReader()
	return c
}
func (c *ConfigSettings) convertToMap(fullPath string) (map[string]interface{}, error) {
	tmp := make(map[string]interface{})
	var err error

	switch reader := c.Reader.(type) {
	case *JSONConfigReader:
		tmp, err = reader.ReadConfigToMap(fullPath)
	case *XMLConfigReader:
		tmp, err = reader.ReadConfigToMap(fullPath)
	case *YAMLConfigReader:
		tmp, err = reader.ReadConfigToMap(fullPath)
	case *TOMLConfigReader:
		tmp, err = reader.ReadConfigToMap(fullPath)
	case *INIConfigReader:
		tmp, err = reader.ReadConfigToMap(fullPath)
	default:
		return nil, fmt.Errorf("unsupported ConfigReader type - %v", reader)
	}

	if err != nil {
		return nil, fmt.Errorf("error converting config to map: %v", err)
	}

	return tmp, nil
}
