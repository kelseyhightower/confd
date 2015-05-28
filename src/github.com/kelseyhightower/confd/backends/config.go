package backends

type Config struct {
	Backend      string
	ClientCaKeys string
	ClientCert   string
	ClientKey    string
	BackendNodes []string
	Scheme       string
	Table        string
	SecKeyFile   string
}
