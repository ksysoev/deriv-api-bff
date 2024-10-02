package handlers

type Config struct {
	Calls []CallConfig `mapstructure:"calls"`
}

type CallConfig struct {
	Method  string            `mapstructure:"method"`
	Params  map[string]string `mapstructure:"params"`
	Backend []BackendConfig   `mapstructure:"backend"`
}

type BackendConfig struct {
	ResponseBody    string   `mapstructure:"response_body"`
	Allow           []string `mapstructure:"allow"`
	RequestTemplate string   `mapstructure:"request_template"`
}
