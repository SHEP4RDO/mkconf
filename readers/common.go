package readers

type ConfigReader interface {
	ReadConfig(filename string, v interface{}) error
	ReadConfigToMap(filename string) (map[string]interface{}, error)
	UpdateConfig(filename string, v interface{}) error
}
