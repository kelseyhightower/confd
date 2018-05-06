package redis

type Config struct {
	Machines  string `mapstructure:"machines"`
	Password  string `mapstructure:"password"`
	Separator string `mapstructure:"separator"`
}
