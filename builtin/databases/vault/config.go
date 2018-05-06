package vault

type Config struct {
	Token    string `mapstructure:"token"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Cert     string `mapstructure:"cert"`
	Key      string `mapstructure:"key"`
	CaCert   string `mapstructure:"caCert"`
	AppID    string `mapstructure:"app-id"`
	UserID   string `mapstructure:"user-id"`
	RoleID   string `mapstructure:"role-id"`
	SecretID string `mapstructure:"secret-id"`
	AuthType string `mapstructure:"authType"`
	Address  string `mapstructure:"address"`
	Path     string `mapstructure:"path"`
}
