package readers

import (
	"fmt"
	"io/ioutil"
	"sync"

	"gopkg.in/ini.v1"
)

// INIConfigReader implements the ConfigReader interface for INI configuration files.
type INIConfigReader struct {
	mu sync.Mutex // Mutex to ensure thread safety during file read and write operations.
}

// ReadConfig reads the content of an INI configuration file into the provided struct.
func (i *INIConfigReader) ReadConfig(filename string, v interface{}) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	cfg, err := ini.Load(fileContent)
	if err != nil {
		return fmt.Errorf("error unmarshalling INI content: %v\n", err)
	}

	if err := cfg.MapTo(&v); err != nil {
		return fmt.Errorf("error unmarshalling INI content: %v\n", err)
	}

	return nil
}

// ReadConfigToMap reads the content of an INI configuration file into a map.
func (i *INIConfigReader) ReadConfigToMap(filename string) (map[string]interface{}, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading INI file: %v\n", err)
	}

	cfg, err := ini.Load(fileContent)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling INI content: %v\n", err)
	}

	configMap := make(map[string]interface{})
	for _, section := range cfg.Sections() {
		sectionMap := make(map[string]interface{})
		for _, key := range section.KeyStrings() {
			sectionMap[key] = section.Key(key).String()
		}
		configMap[section.Name()] = sectionMap
	}

	return configMap, nil
}

// UpdateConfig writes the provided struct as INI to the configuration file.
func (i *INIConfigReader) UpdateConfig(filename string, v interface{}) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	cfg := ini.Empty()
	if err := cfg.ReflectFrom(v); err != nil {
		return fmt.Errorf("error updating INI config: %v", err)
	}

	if err := cfg.SaveTo(filename); err != nil {
		return fmt.Errorf("error writing INI file: %v", err)
	}

	return nil
}
