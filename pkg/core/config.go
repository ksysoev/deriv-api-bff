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
	ResponseBody    string            `mapstructure:"response_body"`
	Allow           []string          `mapstructure:"allow"`
	FieldsMap       map[string]string `mapstructure:"fields_map"`
	RequestTemplate string            `mapstructure:"request_template"`
}
