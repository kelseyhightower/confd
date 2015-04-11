package config

type GlobalConfigFile struct {
	GlobalConfig *GlobalConfig `toml:"global"`
}

type GlobalConfig struct {
	ConfDir  string `toml:"confdir"   cli:"{\"name\":\"confdir\",\"usage\":\"confd conf directory\"}"`
	Interval int    `toml:"interval"  cli:"{\"name\":\"interval\",\"value\":600,\"usage\":\"backend polling interval\"}"`
	Noop     bool   `toml:"noop"      cli:"{\"name\":\"noop\",\"usage\":\"only show pending changes\"}"`
	Onetime  bool   `toml:"onetime"   cli:"{\"name\":\"onetime\",\"usage\":\"run once and exit\"}"`
	Prefix   string `toml:"prefix"    cli:"{\"name\":\"prefix\",\"usage\":\"\"}"`
	Watch    bool   `toml:"watch"     cli:"{\"name\":\"watch\",\"usage\":\"enable watch support\"}"`
	LogLevel string `toml:"log_level" cli:"{\"name\":\"log-level\",\"usage\":\"level which confd should log messages\"}"`
}

func NewGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		Interval: 600,
		Noop:     false,
		Onetime:  false,
		Prefix:   "/",
		Watch:    false,
		LogLevel: "",
	}
}
