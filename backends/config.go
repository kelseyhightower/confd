package backends

type Config struct {
	AuthToken    string
	AuthType     string
	Backend      string
	BasicAuth    bool
	ClientCaKeys string
	ClientCert   string
	ClientKey    string
	BackendNodes []string
	Password     string
	Scheme       string
	Table        string
	Separator    string
	Username     string
	AppID        string
	UserID       string
	RoleID       string
	SecretID     string
	YAMLFile     string
	YAMLBase64   string
	Role         string
}
