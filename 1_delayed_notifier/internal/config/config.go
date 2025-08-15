package config

// Config represents the config
type Config struct {
	Env      string   `yaml:"env" env-default:"local"`
	Server   server   `yaml:"server"`
	RabbitMQ rabbitmq `yaml:"rabbitmq"`
}

type server struct {
	Address string `yaml:"address"`
}

type rabbitmq struct {
	dsn string
}
