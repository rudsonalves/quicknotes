package main

import (
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Hostname   string `env:"HOSTNAME,localhost"`
	ServerPort string `env:"SERVER_PORT,5050"`
	LevelLog   string `env:"LEVEL_LOG,info"`
	DBHost     string `env:"DB_HOST,required"`
	DBPort     string `env:"DB_PORT,required"`
	DBPassword string `env:"DB_PASSWORD,required"`
	DBName     string `env:"DB_NAME,required"`
	DBUser     string `env:"DB_USER,required"`
}

func (c Config) GetLevelLog() slog.Level {
	switch strings.ToLower(c.LevelLog) {
	case "debug":
		return slog.LevelDebug
	case "error":
		return slog.LevelError
	case "warn":
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

func (c Config) DBUrl() string {
	// postgres://DBUser:DBPassword@DBHost:DBPort/DBName
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}

func (c Config) SPrint() (envs string) {
	v := reflect.ValueOf(c)
	t := v.Type()

	for index := 0; index < v.NumField(); index++ {
		field := t.Field(index)
		envTag := strings.Split(field.Tag.Get("env"), ",")
		name := envTag[0]
		value := envTag[1]
		envs += fmt.Sprintf("%s - %s\n", name, value)
	}
	return
}

func (c Config) loadFromEnv() (config Config) {
	v := reflect.ValueOf(c)
	t := v.Type()

	for index := 0; index < v.NumField(); index++ {
		field := t.Field(index)
		envTag := strings.Split(field.Tag.Get("env"), ",")
		envName := envTag[0]
		defaultValue := envTag[1]
		value := os.Getenv(envName)
		if value == "" && defaultValue != "required" {
			f := reflect.ValueOf(&config).Elem().FieldByName(field.Name)
			f.SetString(defaultValue)
		} else {
			f := reflect.ValueOf(&config).Elem().FieldByName(field.Name)
			f.SetString(value)
		}
	}

	return
}

func (c Config) validate() {
	var validateMsg string
	v := reflect.ValueOf(c)
	t := v.Type()

	for index := 0; index < v.NumField(); index++ {
		value := v.Field(index)
		envTag := strings.Split(t.Field(index).Tag.Get("env"), ",")
		envName := envTag[0]
		envValue := envTag[1]
		if envValue == "required" && value.String() == "" {
			validateMsg += fmt.Sprintf("%s is required\n", envName)
		}
	}
	if len(validateMsg) != 0 {
		panic(validateMsg)
	}
}

// Load server values in .env file
func loadConfig() (config Config) {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	config = Config{}
	config = config.loadFromEnv()
	config.validate()

	return
}
