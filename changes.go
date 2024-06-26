package mkconf

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"sync"
	"time"
)

// StartChangeMonitoring initiates monitoring for changes in the specified configuration.
// It sets up a goroutine that periodically checks for configuration changes and triggers notifications.
// The monitoring continues until the associated context is canceled or the quit signal is received.
// Returns an error if the configuration is not found.
func (c *ConfigList) StartChangeMonitoring(configName string, v interface{}) error {
	quit := make(chan struct{})
	settings, ok := c.settings[configName]
	if !ok {
		return fmt.Errorf("config not found: %s", configName)
	}
	c.settings[configName].enableChangeValidation = true
	settings.ctx, settings.cancel = context.WithCancel(context.Background())
	settings.waitGroup.Add(1)

	go func() {
		defer settings.waitGroup.Done()
		mu := &sync.Mutex{}

		for {
			select {
			case <-settings.ch_ChangeValidation:
				close(quit)
				return
			case <-settings.ctx.Done():
				close(quit)
				return
			default:
				err := func() error {
					mu.Lock()
					defer mu.Unlock()

					err := c.checkConfigChanges(configName, v)
					if err != nil {
						fmt.Printf("monitoring: error checking config changes %v : %v\n", configName, err)
						time.Sleep(time.Second * 10)
					}

					return err
				}()

				if err != nil {
					continue
				}

				select {
				case <-time.After(time.Second * time.Duration(settings.checkSec)):
				case <-quit:
					return
				}
			}
		}
	}()
	return nil
}

// StopChangeMonitoring stops the change monitoring for the specified configuration.
// It cancels the associated context, waits for the goroutine to finish, and disables change validation.
func (c *ConfigList) StopChangeMonitoring(configName string) {
	if settings, ok := c.settings[configName]; ok {
		settings.cancel()
		settings.waitGroup.Wait()
		c.settings[configName].enableChangeValidation = false
	}
}

// checkConfigChanges checks for changes in the configuration file and triggers updates accordingly.
// It calculates the hash of the file content, compares it with the last recorded hash,
// and reads the configuration file if a change is detected.
// If change tracking is enabled, it logs the changes.
// Finally, it updates the configuration settings and notifies listeners of the changes.
// Returns an error if there is an issue reading the configuration or calculating the hash.
func (c *ConfigList) checkConfigChanges(configName string, v interface{}) error {
	if c.settings[configName].enableChangeValidation {
		var configMap map[string]interface{}
		var err error
		hash, err := c.settings[configName].calculateFileHash(c.settings[configName].configFullPath)
		if err != nil {
			return err
		}

		c.settings[configName].mu.Lock()
		defer c.settings[configName].mu.Unlock()

		if hash != c.settings[configName].lastConfigHash {
			err := c.settings[configName].Reader.ReadConfig(c.settings[configName].configFullPath, &v)
			if err != nil {
				return err
			}
			if c.settings[configName].enableChangeTracking {
				changes := make([]ConfigChangeLog, 0)
				configMap, err = c.settings[configName].convertToMap(c.settings[configName].configFullPath)
				compareFields(configName, c.settings[configName].configMAP, configMap, &changes)
				c.logChanges(configName, changes)
				if err != nil {
					return fmt.Errorf("monitoring: error v is not of type map[string]interface{}")
				}
			}
			set := c.settings[configName]
			set.config = &v
			set.configMAP = configMap
			set.lastConfigHash = hash
			c.settings[configName] = set

			select {
			case c.settings[configName].Ch_ConfigChanged <- configName:
			case c.settings[configName].Ch_ConfigTracking <- configName:
			}
		}
	}

	return nil
}

// calculateFileHash calculates the MD5 hash of the file content at the specified filename.
// It returns the hexadecimal representation of the hash and an error if there is an issue reading the file.
func (c *ConfigSettings) calculateFileHash(filename string) (string, error) {
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	hash := md5.New()
	_, err = hash.Write(fileContent)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
