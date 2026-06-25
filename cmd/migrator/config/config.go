package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type DataBase struct {
	Host     string `env:"DB_POSTGRE_HOST" env-required:"true"`
	Port     string `env:"DB_POSTGRE_PORT" env-required:"true"`
	User     string `env:"DB_POSTGRE_USER" env-required:"true"`
	Password string `env:"DB_POSTGRE_PASSWORD" env-required:"true"`
	Name     string `env:"DB_POSTGRE_NAME" env-required:"true"`
}

func LoadDatabaseConfig() *DataBase {
	var dbConfig DataBase

	if _, err := os.Stat(".env"); err == nil {
		if err := cleanenv.ReadConfig(".env", &dbConfig); err != nil {
			log.Fatalf("cannot read .env file: %v", err)
		}
	}

	if err := cleanenv.ReadEnv(&dbConfig); err != nil {
		log.Fatalf("cannot read env variables: %v", err)
	}

	return &dbConfig
}
