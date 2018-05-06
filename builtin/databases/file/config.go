package file

type Config struct {
	YamlFile string `mapstructure:"yamlFile"`
	Filter   string `mapstructure:"filter"`
}
