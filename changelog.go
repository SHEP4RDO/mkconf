package mkconf

import (
	"fmt"
	"reflect"
	"time"
)

// ConfigChangeLog represents a log entry capturing changes in configuration fields.
type ConfigChangeLog struct {
	ConfigName string      // Name of the configuration.
	FieldName  string      // Name of the field that changed.
	OldValue   interface{} // Previous value of the field.
	NewValue   interface{} // New value of the field.
	Timestamp  time.Time   // Timestamp of when the change occurred.
}

// compareFields compares two configurations represented as maps and records changes.
// It populates the provided changes slice with ConfigChangeLog entries.
// Returns an error if the oldConfig or newConfig is not a map.
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

// isStruct checks if the given type is a struct.
func isStruct(t reflect.Type) bool {
	return t.Kind() == reflect.Struct
}

// logChanges records the changes in the configuration log for a specific configuration.
// It acquires a lock to ensure thread safety during the log update.
func (c *ConfigList) logChanges(configName string, changes []ConfigChangeLog) {
	c.logMutex.Lock()
	defer c.logMutex.Unlock()
	c.changeLogs[configName] = append(c.changeLogs[configName], changes...)
	c.settings[configName].Ch_ConfigTracking <- configName
}

// GetLogChanges retrieves the log of changes for a specific configuration.
func (c *ConfigList) GetLogChanges(configName string) []ConfigChangeLog {
	return c.changeLogs[configName]
}

// GetChanLogChanges retrieves the channel for tracking changes for a specific configuration.
func (c *ConfigList) GetChanLogChanges(configName string) chan string {
	return c.settings[configName].Ch_ConfigTracking
}
