package config

type (
	Server struct {
		Host string `env:"host"`
		Port int    `env:"port"`
	}

	DB struct {
		Path string `env:"path"`
	}

	Config struct {
		Server      Server      `env:"server"`
		Environment Environment `env:"environment"`
		DB          DB          `env:"db"`
	}
)
