package config

type Config struct {
	Env    string `yaml:"env" env-default:"local"`
	Server server `yaml:"server"`
}

type server struct {
	Address string `yaml:"address"`
}
