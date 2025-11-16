package config

import (
	"fmt"
	"log"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)


type DbConfig struct{
	Host string `env:"DB_HOST"`
	Port int `env:"DB_PORT"`
	User string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	Name string `env:"DB_NAME"`
}

type Config struct{
	Env string `env:"ENV" env-default:"local"`
	Address string `env:"ADDRES" env-default:":8080"`
	Db DbConfig
}


func MustLoad() (*Config) {
	
	err := godotenv.Load("local.env")
	if err != nil {
		fmt.Println(err)
		log.Fatal("unable to load .env file values")		
	}
	
	var cfg Config
	err = cleanenv.ReadEnv(&cfg)
	if err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}

