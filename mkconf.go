package mkconf

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	. "mkconf/readers"
)

type ConfigSettings struct {
	configName     string
	configPath     string
	configFullPath string
	configType     string
	Reader         ConfigReader
	checkSec       int
	repeatSec      int
	lastConfigHash string

	configMAP map[string]interface{}
	config    interface{}

	mu        sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
	waitGroup *sync.WaitGroup

	enableChangeValidation bool
	enableChangeTracking   bool

	ch_ChangeValidation chan struct{}
	Ch_ConfigChanged    chan string
	Ch_ConfigTracking   chan struct{}
}

type ConfigList struct {
	settingsMutex sync.Mutex
	settings      map[string]*ConfigSettings
	changeLogs    map[string][]ConfigChangeLog
	logMutex      sync.Mutex
}

func NewConfigList() *ConfigList {
	list := &ConfigList{}
	list.settings = make(map[string]*ConfigSettings)
	return list
}

func (c *ConfigList) GetSettings(fileName string) *ConfigSettings {
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
func (c *ConfigList) GetChangesChan(configName string) chan string {
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
		return fmt.Errorf("load config %v: error while read config: %v\n", configName, err)
	}
	c.settings[configName].config = v
	return nil
}

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

func (s *ConfigSettings) checkReader() ConfigReader {
	_type := strings.ToLower(s.configType)
	switch _type {
	case ".json", ".mk.json":
		return &JSONConfigReader{}
	case ".xml", ".mk.xml":
		return &XMLConfigReader{}
	case ".yaml", ".yml", ".mk.yaml", ".mk.yml":
		return &YAMLConfigReader{}
	case ".toml", ".mk.toml":
		return &TOMLConfigReader{}
	case ".ini", ".mk.ini":
		return &INIConfigReader{}
	default:
		return nil
	}
}

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
		return fmt.Errorf("mkconf: error add new config %v: %v", configName, err)
	}
	return nil
}

func (c *ConfigSettings) defineHash(configName string, v interface{}) error {
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
