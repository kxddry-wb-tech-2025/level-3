package config

type Config struct {
	Env     string  `yaml:"env" env-default:"local"`
	Server  server  `yaml:"server"`
	Storage storage `yaml:"storage"`
}

type server struct {
	Address string `yaml:"address"`
}

type storage struct {
	dsn string
}
