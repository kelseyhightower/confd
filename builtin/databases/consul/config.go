package consul

type Config struct {
	Nodes  string `mapstructure:"nodes"`
	Scheme string `mapstructure:"scheme"`
	Cert   string `mapstructure:"cert"`
	Key    string `mapstructure:"key"`
	CaCert string `mapstructure:"caCert"`
}
