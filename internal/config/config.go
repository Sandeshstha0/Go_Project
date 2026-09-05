package config

import (
	"flag"
	"log"
	"os"
)

type HTTPServer struct {
	Addr string `yaml:"addr" env:"ADDR" env-required:"true"`
}

// env_default:"production"

type Config struct {
	Env         string `yaml:"env" env:"ENV" env-required:"true" ` //struct tags
	StoragePath string `yaml:"storage_path" env:"STORAGE_PATH" env-required:"true"`
	HTTPServer  `yaml:"http_server"`
}

func MustLoad() {
	var ConfigPath string

	ConfigPath = os.Getenv("CONFIG_PATH")
	if ConfigPath == "" {
		configPath := flag.String("config", "", "path to the configuration file")
		flag.Parse()
		ConfigPath = *configPath
	}

	if ConfigPath == "" {
		log.Fatal("config path is not set: use CONFIG_PATH env var or --config flag")
	}
}