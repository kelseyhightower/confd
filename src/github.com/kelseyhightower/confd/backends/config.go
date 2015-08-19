package backends

type Config struct {
	AuthToken    string
	Backend      string
	ClientCaKeys string
	ClientCert   string
	ClientKey    string
	BackendNodes []string
	NoDiscover   bool
	Scheme       string
	Table        string
}
