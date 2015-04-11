package config

type BackendConfigFile interface {
	ConfigFile() BackendConfig
}

func NewBackendConfigFile(backend string) (bcf BackendConfigFile) {
	switch backend {
	case "consul":
		bcf = &ConsulBackendConfigFile{NewConsulBackendConfig()}
	case "env":
		bcf = &EnvBackendConfigFile{NewEnvBackendConfig()}
	case "etcd":
		bcf = &EtcdBackendConfigFile{NewEtcdBackendConfig()}
	case "redis":
		bcf = &RedisBackendConfigFile{NewRedisBackendConfig()}
	case "zookeeper":
		bcf = &ZookeeperBackendConfigFile{NewZookeeperBackendConfig()}
	case "dynamodb":
		bcf = &DynamoDBBackendConfigFile{NewDynamoDBBackendConfig()}
	case "fs":
		bcf = &FsBackendConfigFile{NewFsBackendConfig()}
	default:
		panic("invalid backend, this should never happen!")
	}
	return
}

type BackendConfig interface {
	Type() string
	IsWatchSupported() bool
}

func NewBackendConfig(backend string) BackendConfig {
	return NewBackendConfigFile(backend).ConfigFile()
}

type ConsulBackendConfigFile struct {
	ConsulBackendConfig *ConsulBackendConfig `toml:"consul"`
}

func (c *ConsulBackendConfigFile) ConfigFile() BackendConfig {
	return c.ConsulBackendConfig
}

type ConsulBackendConfig struct {
	Scheme  string   `toml:"scheme"          cli:"{\"name\":\"scheme\"}"`
	Nodes   []string `toml:"nodes"           cli:"{\"name\":\"node\"}"`
	Srv     string   `toml:"srv"             cli:"{\"name\":\"srv\"}"`
	Cert    string   `toml:"client_cert"     cli:"{\"name\":\"client-cert\",\"envvar\":\"CONFD_CLIENT_CERT\"}"`
	CertKey string   `toml:"client_cert_key" cli:"{\"name\":\"client-cert-key\",\"envvar\":\"CONFD_CLIENT_CERT_KEY\"}"`
	CAKeys  string   `toml:"client_ca_keys"  cli:"{\"name\":\"client-ca-keys\",\"envvar\":\"CONFD_CLIENT_CA_KEYS\"}"`
}

func NewConsulBackendConfig() *ConsulBackendConfig {
	return &ConsulBackendConfig{
		Scheme:  "http",
		Nodes:   []string{"127.0.0.1:8500"},
		Srv:     "",
		Cert:    "",
		CertKey: "",
		CAKeys:  "",
	}
}

func (*ConsulBackendConfig) Type() string {
	return "consul"
}

func (*ConsulBackendConfig) IsWatchSupported() bool {
	return true
}

type EnvBackendConfigFile struct {
	EnvBackendConfig *EnvBackendConfig `toml:"env"`
}

func (e *EnvBackendConfigFile) ConfigFile() BackendConfig {
	return e.EnvBackendConfig
}

type EnvBackendConfig struct {
}

func NewEnvBackendConfig() *EnvBackendConfig {
	return &EnvBackendConfig{}
}

func (*EnvBackendConfig) Type() string {
	return "env"
}

func (*EnvBackendConfig) IsWatchSupported() bool {
	return true
}

type EtcdBackendConfigFile struct {
	EtcdBackendConfig *EtcdBackendConfig `toml:"etcd"`
}

func (e *EtcdBackendConfigFile) ConfigFile() BackendConfig {
	return e.EtcdBackendConfig
}

type EtcdBackendConfig struct {
	Nodes   []string `toml:"nodes"           cli:"{\"name\":\"node\",\"envvar\":\"ETCDCTL_PEERS\"}"`
	Srv     string   `toml:"srv"             cli:"{\"name\":\"srv\"}"`
	Cert    string   `toml:"client_cert"     cli:"{\"name\":\"client-cert\",\"envvar\":\"CONFD_CLIENT_CERT\"}"`
	CertKey string   `toml:"client_cert_key" cli:"{\"name\":\"client-cert-key\",\"envvar\":\"CONFD_CLIENT_CERT_KEY\"}"`
	CAKeys  string   `toml:"client_ca_keys"  cli:"{\"name\":\"client-ca-keys\",\"envvar\":\"CONFD_CLIENT_CA_KEYS\"}"`
}

func NewEtcdBackendConfig() *EtcdBackendConfig {
	return &EtcdBackendConfig{
		Nodes:   []string{"http://127.0.0.1:2379"},
		Srv:     "",
		Cert:    "",
		CertKey: "",
		CAKeys:  "",
	}
}

func (*EtcdBackendConfig) Type() string {
	return "etcd"
}

func (*EtcdBackendConfig) IsWatchSupported() bool {
	return true
}

type RedisBackendConfigFile struct {
	RedisBackendConfig *RedisBackendConfig `toml:"redis"`
}

func (r *RedisBackendConfigFile) ConfigFile() BackendConfig {
	return r.RedisBackendConfig
}

type RedisBackendConfig struct {
	Nodes []string `toml:"nodes" cli:"{\"name\":\"node\"}"`
	Srv   string   `toml:"srv"   cli:"{\"name\":\"srv\"}"`
}

func NewRedisBackendConfig() *RedisBackendConfig {
	return &RedisBackendConfig{
		Nodes: []string{"127.0.0.1:6379"},
		Srv:   "",
	}
}

func (*RedisBackendConfig) Type() string {
	return "redis"
}

func (*RedisBackendConfig) IsWatchSupported() bool {
	return false
}

type ZookeeperBackendConfigFile struct {
	ZookeeperBackendConfig *ZookeeperBackendConfig `toml:"zookeeper"`
}

func (z *ZookeeperBackendConfigFile) ConfigFile() BackendConfig {
	return z.ZookeeperBackendConfig
}

type ZookeeperBackendConfig struct {
	Nodes []string `toml:"nodes" cli:"{\"name\":\"node\"}"`
	Srv   string   `toml:"srv"   cli:"{\"name\":\"srv\"}"`
}

func NewZookeeperBackendConfig() *ZookeeperBackendConfig {
	return &ZookeeperBackendConfig{
		Nodes: []string{"127.0.0.1:2181"},
		Srv:   "",
	}
}

func (*ZookeeperBackendConfig) Type() string {
	return "zookeeper"
}

func (*ZookeeperBackendConfig) IsWatchSupported() bool {
	return false
}

type DynamoDBBackendConfigFile struct {
	DynamoDBBackendConfig *DynamoDBBackendConfig `toml:"dynamodb"`
}

func (f *DynamoDBBackendConfigFile) ConfigFile() BackendConfig {
	return f.DynamoDBBackendConfig
}

type DynamoDBBackendConfig struct {
	Table string `toml:"table" cli:"{\"name\":\"table\"}"`
}

func NewDynamoDBBackendConfig() *DynamoDBBackendConfig {
	return &DynamoDBBackendConfig{
		Table: "/",
	}
}

func (*DynamoDBBackendConfig) Type() string {
	return "dynamodb"
}

func (*DynamoDBBackendConfig) IsWatchSupported() bool {
	return false
}

type FsBackendConfigFile struct {
	FsBackendConfig *FsBackendConfig `toml:"fs"`
}

func (f *FsBackendConfigFile) ConfigFile() BackendConfig {
	return f.FsBackendConfig
}

type FsBackendConfig struct {
	RootPath    string `toml:"root_path"     cli:"{\"name\":\"root-path\"}"`
	MaxFileSize int    `toml:"max_file_size" cli:"{\"name\":\"max-file-size\",\"value\":1048576}"`
}

func NewFsBackendConfig() *FsBackendConfig {
	return &FsBackendConfig{
		RootPath:    "/",
		MaxFileSize: 1048576,
	}
}

func (*FsBackendConfig) Type() string {
	return "fs"
}

func (*FsBackendConfig) IsWatchSupported() bool {
	return false
}
