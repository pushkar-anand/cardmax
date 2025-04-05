package config

type (
	Server struct {
		Host string `env:"host"`
		Port int    `env:"port"`
	}

	Config struct {
		Server      Server      `env:"server"`
		Environment Environment `env:"environment"`
	}
)
