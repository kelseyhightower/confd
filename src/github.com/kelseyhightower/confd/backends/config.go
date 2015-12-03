package backends

type Config struct {
	AuthToken    string
	Backend      string
	ClientCaKeys string
	ClientCert   string
	ClientKey    string
	BackendNodes []string
	NoSync       bool
	Scheme       string
	Table        string
}
