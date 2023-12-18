package readers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
)

type JSONConfigReader struct {
	mu sync.Mutex
}

func (j *JSONConfigReader) ReadConfig(filename string, v interface{}) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading JSON file: %v\n", err)
	}

	if err := json.Unmarshal(fileContent, &v); err != nil {
		return fmt.Errorf("error unmarshalling JSON content: %v\n", err)
	}

	return nil
}

func (j *JSONConfigReader) ReadConfigToMap(filename string) (map[string]interface{}, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading JSON file: %v\n", err)
	}

	var configMap map[string]interface{}
	if err := json.Unmarshal(fileContent, &configMap); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON content: %v\n", err)
	}

	return configMap, nil
}

func (j *JSONConfigReader) UpdateConfig(filename string, v interface{}) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON content: %v", err)
	}
	err = ioutil.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing JSON file: %v", err)
	}

	return nil
}
