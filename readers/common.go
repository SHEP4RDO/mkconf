package readers

// ConfigReader is an interface for reading and updating configuration files.
type ConfigReader interface {
	ReadConfig(filename string, v interface{}) error                 // ReadConfig reads the content of a configuration file into the provided struct.
	ReadConfigToMap(filename string) (map[string]interface{}, error) // ReadConfigToMap reads the content of a configuration file into a map.
	UpdateConfig(filename string, v interface{}) error               // UpdateConfig writes the provided struct as JSON to the configuration file.
}
