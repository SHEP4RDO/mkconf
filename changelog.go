package mkconf

import (
	"fmt"
	"reflect"
	"time"
)

type ConfigChangeLog struct {
	ConfigName string
	FieldName  string
	OldValue   interface{}
	NewValue   interface{}
	Timestamp  time.Time
}

func compareFields(configName, configFullName string, oldConfig, newConfig interface{}, changes *[]ConfigChangeLog) error {
	oldMap, ok := oldConfig.(map[string]interface{})
	if !ok {
		return fmt.Errorf("monitoring changes: error while check changes %v : oldConfig is not of type map[string]interface{}", configName)
	}

	newMap, ok := newConfig.(map[string]interface{})
	if !ok {
		return fmt.Errorf("monitoring changes: error while check changes %v : newConfig is not of type map[string]interface{}", configName)
	}

	for key, oldValue := range oldMap {
		newValue, exists := newMap[key]
		if exists {
			if !reflect.DeepEqual(oldValue, newValue) {
				changeLog := ConfigChangeLog{
					ConfigName: configName,
					FieldName:  key,
					OldValue:   oldValue,
					NewValue:   newValue,
					Timestamp:  time.Now(),
				}
				*changes = append(*changes, changeLog)
			}
		} else {
			changeLog := ConfigChangeLog{
				ConfigName: configName,
				FieldName:  key,
				OldValue:   oldValue,
				NewValue:   nil,
				Timestamp:  time.Now(),
			}
			*changes = append(*changes, changeLog)
		}
	}

	for key, newValue := range newMap {
		_, exists := oldMap[key]
		if !exists {
			changeLog := ConfigChangeLog{
				ConfigName: configName,
				FieldName:  key,
				OldValue:   nil,
				NewValue:   newValue,
				Timestamp:  time.Now(),
			}
			*changes = append(*changes, changeLog)
		}
	}

	return nil
}

func isStruct(t reflect.Type) bool {
	return t.Kind() == reflect.Struct
}

func (c *ConfigList) logChanges(configName string, changes []ConfigChangeLog) {
	c.logMutex.Lock()
	defer c.logMutex.Unlock()
	c.changeLogs[configName] = append(c.changeLogs[configName], changes...)
	c.settings[configName].Ch_ConfigTracking <- struct{}{}
}

func (c *ConfigList) GetLogChanges(configName string) []ConfigChangeLog {
	return c.changeLogs[configName]
}
func (c *ConfigList) GetChanLogChanges(configName string) chan struct{} {
	return c.settings[configName].Ch_ConfigTracking
}
