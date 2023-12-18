package readers

import (
	"fmt"
	"io/ioutil"
	"sync"

	"gopkg.in/ini.v1"
)

type INIConfigReader struct {
	mu sync.Mutex
}

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
