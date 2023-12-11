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

func (c *ConfigManager) StartChangeMonitoring(configName string, v interface{}) error {
	quit := make(chan struct{})
	settings, ok := c.settings[configName]
	if !ok {
		return fmt.Errorf("config not found: %s", configName)
	}

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
						fmt.Printf("monitoring: error checking config changes: %v\n", err)
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
func (c *ConfigManager) StopChangeMonitoring(configName string) {
	if settings, ok := c.settings[configName]; ok {
		settings.cancel()
		settings.waitGroup.Wait()
	}
}
func (c *ConfigManager) checkConfigChanges(configName string, v interface{}) error {
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
				compareFields(configName, c.settings[configName].configName+c.settings[configName].configType, c.settings[configName].configMAP, configMap, &changes)
				c.logChanges(configName, changes)

				if err != nil {
					return fmt.Errorf("monitoring: error v is not of type map[string]interface{}")
				}
			}
			set := c.settings[configName]
			set.config = v
			set.configMAP = configMap
			c.settings[configName] = set

			select {
			case c.settings[configName].Ch_ConfigChanged <- struct{}{}:
			case c.settings[configName].Ch_ConfigTracking <- struct{}{}:
			}
		}
	}

	return nil
}

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
