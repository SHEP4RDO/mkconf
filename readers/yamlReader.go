package readers

import (
	"fmt"
	"io/ioutil"
	"sync"

	"gopkg.in/yaml.v2"
)

type YAMLConfigReader struct {
	mu sync.Mutex
}

func (y *YAMLConfigReader) ReadConfig(filename string, v interface{}) error {
	y.mu.Lock()
	defer y.mu.Unlock()
	yamlContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading YAML file: %v\n", err)
	}

	if err := yaml.Unmarshal(yamlContent, v); err != nil {
		return fmt.Errorf("error unmarshalling YAML content: %v\n", err)
	}

	return nil
}

func (y *YAMLConfigReader) ReadConfigToMap(filename string) (map[string]interface{}, error) {
	y.mu.Lock()
	defer y.mu.Unlock()
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading YAML content: %v\n", err)
	}

	var configMap map[string]interface{}
	if err := yaml.Unmarshal(fileContent, &configMap); err != nil {
		return nil, fmt.Errorf("error unmarshalling YAML content: %v\n", err)
	}

	return configMap, nil
}

func (y *YAMLConfigReader) UpdateConfig(filename string, v interface{}) error {
	y.mu.Lock()
	defer y.mu.Unlock()
	yamlData, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("error marshalling YAML: %v", err)
	}

	if err := ioutil.WriteFile(filename, yamlData, 0644); err != nil {
		return fmt.Errorf("error writing YAML file: %v", err)
	}

	return nil
}
