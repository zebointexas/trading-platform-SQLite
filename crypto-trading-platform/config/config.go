package config

import (
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	Name     string `yaml:"Name"`
	RestConf rest.RestConf
	Database struct {
		Path string
	}
	Kraken struct {
		APIKey    string
		APISecret string
	}
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
}

 