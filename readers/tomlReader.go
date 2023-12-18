package readers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/pelletier/go-toml"
)

type TOMLConfigReader struct {
	mu sync.Mutex
}

func (t *TOMLConfigReader) ReadConfig(filename string, v interface{}) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading TOML content: %v\n", err)
	}

	tree, err := toml.Load(string(fileContent))
	if err != nil {
		return fmt.Errorf("error unmarshalling TOML content: %v\n", err)
	}

	if err := tree.Unmarshal(&v); err != nil {
		return fmt.Errorf("error unmarshalling TOML content: %v\n", err)
	}

	return nil
}
func (t *TOMLConfigReader) ReadConfigToMap(filename string) (map[string]interface{}, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading TOML content: %v\n", err)
	}

	var configMap map[string]interface{}
	tree, err := toml.Load(string(fileContent))
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling TOML content: %v\n", err)
	}

	tree.Unmarshal(&configMap)

	return configMap, nil
}

func (t *TOMLConfigReader) UpdateConfig(filename string, v interface{}) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(v); err != nil {
		return fmt.Errorf("error encoding TOML: %v", err)
	}

	if err := ioutil.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("error writing TOML file: %v", err)
	}

	return nil
}
