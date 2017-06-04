package backends

type Config struct {
	AuthToken       string
	AuthType        string
	Backend         string
	BackendFallback string
	BackendNodes    []string
	BasicAuth       bool
	ClientCaKeys    string
	ClientCert      string
	ClientKey       string
	Endpoint        string
	Password        string
	Scheme          string
	StorageAccount  string
	Table           string
	Username        string
	AppID           string
	UserID          string
	YAMLFile        string
}
