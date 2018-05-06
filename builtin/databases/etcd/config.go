package etcd

type Config struct {
	Machines  string `mapstructure:"machines"`
	BasicAuth string `mapstructure:"basicAuth"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	Cert      string `mapstructure:"cert"`
	Key       string `mapstructure:"key"`
	CaCert    string `mapstructure:"caCert"`
}
