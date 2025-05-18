package config

func InitConfig() *Config {
	// read config from file
	return &Config{}
}

type Config struct {
	Database DatabaseConfig
}
type DatabaseConfig struct {
	Host     string
	Username string
	Password string
}
