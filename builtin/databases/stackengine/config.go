package stackengine

type Config struct {
	Nodes     string `mapstructure:"nodes"`
	AuthToken string `mapstructure:"authToken"`
	Scheme    string `mapstructure:"scheme"`
	Cert      string `mapstructure:"cert"`
	Key       string `mapstructure:"key"`
	CaCert    string `mapstructure:"caCert"`
}
