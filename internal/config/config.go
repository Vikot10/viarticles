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
	AccessToken string `env:"ACCESS_TOKEN"`
}

type Telegram struct {
	BotToken string `env:"BOT_TOKEN"`
}

type Habr struct {
	AuthToken string `env:"AUTH_TOKEN"`
}

type Config struct {
	IsDebug  bool     `env:"IS_DEBUG"`
	Postgres Postgres `env:"POSTGRES"`
	Address  string   `env:"ADDRESS"`
	Vk       Vk       `env:"VK"`
	Telegram Telegram `env:"TELEGRAM"`
	Habr     Habr     `env:"HABR"`
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
