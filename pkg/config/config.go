package config

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	HttpServerPort string `env:"HTTP_SERVER_PORT" envDefault:"8080"`

	MongoURI    string `env:"MONGO_URI,required"`
	MongoDBName string `env:"MONGO_DB_NAME,required"`

	FilePath string `env:"FILE_PATH,required"`
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, fmt.Errorf("could not parse config: %w", err)
	}

	return cfg, nil
}
