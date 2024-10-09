package core

type Config struct {
	Calls []CallConfig `mapstructure:"calls"`
}

type CallConfig struct {
	Method  string            `mapstructure:"method"`
	Params  map[string]string `mapstructure:"params"`
	Backend []BackendConfig   `mapstructure:"backend"`
}

type BackendConfig struct {
	FieldsMap       map[string]string `mapstructure:"fields_map"`
	ResponseBody    string            `mapstructure:"response_body"`
	RequestTemplate string            `mapstructure:"request_template"`
	Allow           []string          `mapstructure:"allow"`
}
