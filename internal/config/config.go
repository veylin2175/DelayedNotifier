package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env        string     `yaml:"env" env-default:"local"`
	Database   Database   `yaml:"database"`
	HTTPServer HTTPServer `yaml:"http_server"`
	Redis      Redis      `yaml:"redis"`
	Rabbit     Rabbit     `yaml:"rabbit"`
	TGToken    string     `yaml:"tg_token"`
}

type Database struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     int    `yaml:"port" env-default:"5432"`
	User     string `yaml:"user" env-default:"postgres"`
	Password string `yaml:"password" env-required:"true"`
	DBName   string `yaml:"dbname" env-required:"true"`
	SSLMode  string `yaml:"sslmode" env-default:"disable"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8036"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type Redis struct {
	Addr         string        `yaml:"addr" env-default:"localhost:6379"`
	Password     string        `yaml:"password" env-default:""`
	DB           int           `yaml:"db" env-default:"0"`
	DialTimeout  time.Duration `yaml:"dial_timeout" env-default:"5s"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env-default:"3s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env-default:"3s"`
	PoolSize     int           `yaml:"pool_size" env-default:"10"`
	PoolTimeout  time.Duration `yaml:"pool_timeout" env-default:"30s"`
}

type Rabbit struct {
	Host      string `yaml:"host" env-default:"localhost"`
	Port      int    `yaml:"port" env-default:"5672"`
	User      string `yaml:"user" env-default:"guest"`
	Password  string `yaml:"password" env-default:"guest"`
	QueueName string `yaml:"queue_name" env-default:"notifications_queue"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}

	return MustLoadByPath(path)
}

func MustLoadByPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	return &cfg
}

// fetchConfigPath извлекает путь конфигурации из флага командной строки или переменной среды.
// Приоритет: flag > env > default.
// Дефолтное значение — пустая строка.
func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "config file path")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
