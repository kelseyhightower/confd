package backends

type Config struct {
	AuthToken    string
	Backend      string
	BasicAuth    bool
	ClientCaKeys string
	ClientCert   string
	ClientKey    string
	BackendNodes []string
	Password     string
	Scheme       string
	Table        string
	Username     string
}
