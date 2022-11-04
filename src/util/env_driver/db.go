package env_driver

import (
	"os"

	"github.com/joho/godotenv"
)

type PostgreEnv struct {
	Username string
	Password string
	Host     string
	Port     string
	Database string
	Schema   string
}

func ReadPostgreEnv() (PostgreEnv, error) {
	if err := godotenv.Load(".env"); err != nil {
		return PostgreEnv{}, err
	}

	return PostgreEnv{
		Username: os.Getenv("GORM_USERNAME"),
		Password: os.Getenv("GORM_PASSWORD"),
		Host:     os.Getenv("GORM_HOST"),
		Port:     os.Getenv("GORM_PORT"),
		Database: os.Getenv("GORM_DATABASE"),
		Schema:   os.Getenv("GORM_SCHEMA"),
	}, nil
}
