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
	DBPassword   string `env:"POSTGRES_PASSWORD,required"`
	DBUser       string `env:"POSTGRES_USER,required"`
	ServerPort   string `env:"QNS_SERVER_PORT,5000"`
	LevelLog     string `env:"QNS_LEVEL_LOG,info"`
	DBHost       string `env:"QNS_DB_HOST,required"`
	DBPort       string `env:"QNS_DB_PORT,required"`
	DBName       string `env:"QNS_DB_NAME,required"`
	MailHost     string `env:"QNS_SMTP_HOST,required"`
	MailPort     string `env:"QNS_SMTP_PORT,required"`
	MailUserName string `env:"QNS_SMTP_USER_NAME,required"`
	MailUserPass string `env:"QNS_SMTP_USER_PASS,required"`
	MailFrom     string `env:"QNS_SMTP_FROM,nao-responder@quick.com"`
	CSRFKey      string `env:"QNS_CSRF_KEY,required"`
}

func (cfg Config) GetLevelLog() slog.Level {
	switch strings.ToLower(cfg.LevelLog) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "error":
		return slog.LevelError
	case "warn":
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

func (c Config) DBConnURL() string {
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

func (cfg Config) SPrint() (envs string) {
	v := reflect.ValueOf(cfg)
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

func (cfg Config) loadFromEnv() (config Config) {
	v := reflect.ValueOf(cfg)
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

func (cfg Config) validate() {
	var validateMsg string
	v := reflect.ValueOf(cfg)
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
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	config = config.loadFromEnv()
	config.validate()

	return
}
