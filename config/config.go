package config

import (
	"fmt"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string     `yaml:"env" env-default:"development"`
	DB         DBconf     `yaml:"db" end-required:"true"`
	HTTPServer HTTPServer `yaml:"httpserver"`
}

type DBconf struct {
	User     string `yaml:"user"`
	DBName   string `yaml:"dbname"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Sslmode  string `yaml:"sslmode"`
}

type HTTPServer struct {
	Address string `yaml:"address" env-default:"0.0.0.0:8080"`
}

func LoadConfig() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is empty!")
	}

	if _, err := os.Stat(configPath); err != nil {
		log.Fatalf("error openning config file: %s", err)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatal("error reading config file: %s", err)
	}

	return &cfg
}

func MakeBDPath(cfg Config) string {
	result := fmt.Sprintf("user=%s dbname=%s password=%s host=%s sslmode=%s", cfg.DB.User, cfg.DB.DBName, cfg.DB.Password, cfg.DB.Host, cfg.DB.Sslmode)
	return result
}
