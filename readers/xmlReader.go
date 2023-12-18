package readers

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"sync"
)

type XMLConfigReader struct {
	mu sync.Mutex
}

func (x *XMLConfigReader) ReadConfig(filename string, v interface{}) error {
	x.mu.Lock()
	defer x.mu.Unlock()
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error unmarshalling XML content: %v\n", err)
	}

	if err := xml.Unmarshal(fileContent, &v); err != nil {
		return fmt.Errorf("error unmarshalling XML content: %v\n", err)
	}

	return nil
}
func (x *XMLConfigReader) ReadConfigToMap(filename string) (map[string]interface{}, error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading XML file: %v\n", err)
	}

	var configMap map[string]interface{}
	if err := xml.Unmarshal(fileContent, &configMap); err != nil {
		return nil, fmt.Errorf("error unmarshalling XML content: %v\n", err)
	}

	return configMap, nil
}

func (x *XMLConfigReader) UpdateConfig(filename string, v interface{}) error {
	x.mu.Lock()
	defer x.mu.Unlock()
	xmlData, err := xml.MarshalIndent(v, "", "    ")
	if err != nil {
		return fmt.Errorf("error marshalling XML: %v", err)
	}

	if err := ioutil.WriteFile(filename, xmlData, 0644); err != nil {
		return fmt.Errorf("error writing XML file: %v", err)
	}

	return nil
}
