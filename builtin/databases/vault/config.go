package vault

type Config struct {
	Token    string `mapstructure:"token"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Cert     string `mapstructure:"cert"`
	Key      string `mapstructure:"key"`
	CaCert   string `mapstructure:"caCert"`
	AppId    string `mapstructure:"app-id"`
	UserId   string `mapstructure:"user-id"`
	AuthType string `mapstructure:"authType"`
	Address  string `mapstructure:"address"`
}
