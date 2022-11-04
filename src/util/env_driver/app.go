package env_driver

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type AppEnv struct {
	Port    string
	TimeOut time.Duration
}

type Env struct {
	App     AppEnv
	Postgre PostgreEnv
	Email   string
}

func ReadAppEnv() (AppEnv, error) {
	if err := godotenv.Load(".env"); err != nil {
		return AppEnv{}, err
	}

	timeOutInt, err := strconv.Atoi(os.Getenv("APP_TIMEOUT"))
	if err != nil {
		return AppEnv{}, err
	}

	timeOut := time.Duration(timeOutInt) * time.Second

	return AppEnv{
		Port:    os.Getenv("PORT"),
		TimeOut: timeOut,
	}, nil
}

func NewEnvDriver() (Env, error) {
	appEnv, err := ReadAppEnv()
	if err != nil {
		return Env{}, err
	}

	postgreEnv, err := ReadPostgreEnv()
	if err != nil {
		return Env{}, err
	}
	return Env{
		App:     appEnv,
		Postgre: postgreEnv,
	}, nil
}
