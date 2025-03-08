package config

import (
	"github.com/cristalhq/aconfig"
)

type Postgres struct {
	Host        string `env:"HOST"`
	Port        int    `env:"PORT"`
	Username    string `env:"USERNAME"`
	Password    string `env:"PASSWORD"`
	Database    string `env:"DATABASE"`
	SSLMode     string `env:"SSL_MODE" default:"disable"`
	SSLCertPath string `env:"SSL_CERT_PATH"`
	NeedMigrate bool   `env:"NEED_MIGRATE" default:"false"`
}

type Vk struct {
	AccessToken string `env:"VK_ACCESS_TOKEN"`
}

type Config struct {
	IsDebug  bool     `env:"IS_DEBUG"`
	Postgres Postgres `env:"POSTGRES"`
	Address  string   `env:"ADDRESS"`
	Vk       Vk       `env:"VK"`
}

func Load() *Config {
	cfg := Config{}

	err := aconfig.LoaderFor(&cfg, aconfig.Config{
		EnvPrefix: "VIARTICLES",
	}).Load()
	if err != nil {
		panic(err)
	}

	return &cfg
}
