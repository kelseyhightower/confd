package config

type TemplateConfigFile struct {
	TemplateConfig TemplateConfig `toml:"template"`
}

type TemplateConfig struct {
	Src           string   `toml:"src"             cli:"{\"name\":\"tmpl-src\"}"`
	Dest          string   `toml:"dest"            cli:"{\"name\":\"tmpl-dest\"}"`
	Uid           int      `toml:"uid"             cli:"{\"name\":\"tmpl-uid\"}"`
	Gid           int      `toml:"gid"             cli:"{\"name\":\"tmpl-gid\"}"`
	Mode          string   `toml:"mode"            cli:"{\"name\":\"tmpl-mode\",\"value\":\"0644\"}"`
	KeepStageFile bool     `toml:"keep_stage_file" cli:"{\"name\":\"tmpl-keep-stage-file\"}"`
	Prefix        string   `toml:"prefix"`
	Keys          []string `toml:"keys"            cli:"{\"name\":\"tmpl-key\"}"`
	CheckCmd      string   `toml:"check_cmd"       cli:"{\"name\":\"tmpl-check-cmd\"}"`
	ReloadCmd     string   `toml:"reload_cmd"      cli:"{\"name\":\"tmpl-reload-cmd\"}"`
}

func NewTemplateConfig() *TemplateConfig {
	return &TemplateConfig{
		Src:           "",
		Dest:          "",
		Uid:           0,
		Gid:           0,
		Mode:          "0644",
		KeepStageFile: false,
		Prefix:        "/",
		Keys:          []string{"/"},
		CheckCmd:      "",
		ReloadCmd:     "",
	}
}
