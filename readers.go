package mkconf

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"

	"github.com/pelletier/go-toml"
	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v2"
)

type ConfigReader interface {
	ReadConfig(filename string, v interface{}) error
	ReadConfigToMap(filename string) (map[string]interface{}, error)
}

type JSONConfigReader struct{}

func (j *JSONConfigReader) ReadConfig(filename string, v interface{}) error {
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(fileContent, v); err != nil {
		return err
	}

	return nil
}

func (j *JSONConfigReader) ReadConfigToMap(filename string) (map[string]interface{}, error) {
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var configMap map[string]interface{}
	if err := json.Unmarshal(fileContent, &configMap); err != nil {
		return nil, err
	}

	return configMap, nil
}

type XMLConfigReader struct{}

func (x *XMLConfigReader) ReadConfig(filename string, v interface{}) error {
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	if err := xml.Unmarshal(fileContent, v); err != nil {
		return err
	}

	return nil
}
func (x *XMLConfigReader) ReadConfigToMap(filename string) (map[string]interface{}, error) {
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var configMap map[string]interface{}
	if err := xml.Unmarshal(fileContent, &configMap); err != nil {
		return nil, err
	}

	return configMap, nil
}

type YAMLConfigReader struct{}

func (y *YAMLConfigReader) ReadConfig(filename string, v interface{}) error {
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(fileContent, v); err != nil {
		return err
	}

	return nil
}
func (y *YAMLConfigReader) ReadConfigToMap(filename string) (map[string]interface{}, error) {
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var configMap map[string]interface{}
	if err := yaml.Unmarshal(fileContent, &configMap); err != nil {
		return nil, err
	}

	return configMap, nil
}

type TOMLConfigReader struct{}

func (t *TOMLConfigReader) ReadConfig(filename string, v interface{}) error {
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	tree, err := toml.Load(string(fileContent))
	if err != nil {
		return err
	}

	if err := tree.Unmarshal(v); err != nil {
		return err
	}

	return nil
}
func (t *TOMLConfigReader) ReadConfigToMap(filename string) (map[string]interface{}, error) {
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var configMap map[string]interface{}
	tree, err := toml.Load(string(fileContent))
	if err != nil {
		return nil, err
	}

	tree.Unmarshal(&configMap)

	return configMap, nil
}

type INIConfigReader struct{}

func (i *INIConfigReader) ReadConfig(filename string, v interface{}) error {
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	cfg, err := ini.Load(fileContent)
	if err != nil {
		return err
	}

	if err := cfg.MapTo(v); err != nil {
		return err
	}

	return nil
}

func (i *INIConfigReader) ReadConfigToMap(filename string) (map[string]interface{}, error) {
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	cfg, err := ini.Load(fileContent)
	if err != nil {
		return nil, err
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
