package config

import (
	"errors"
	"github.com/jackc/pgx"
	"github.com/pelletier/go-toml"
	"log/slog"
	"os"
)

var Conn pgx.ConnConfig
var Server ServerConfig

type ServerConfig struct {
	Host string
	Port uint16
}

func loadConfig(Path string) (pgx.ConnConfig, ServerConfig, error) {
	content, err := os.ReadFile(Path)
	if err != nil {
		return pgx.ConnConfig{}, ServerConfig{}, errors.New("failed to read config: " + err.Error())
	}
	config, err := toml.Load(string(content))
	if err != nil {
		return pgx.ConnConfig{}, ServerConfig{}, errors.New("failed to load config: " + err.Error())
	}
	return pgx.ConnConfig{
			Host:     config.Get("database.host").(string),
			Port:     uint16(config.Get("database.port").(int64)),
			Database: config.Get("database.database").(string),
			User:     config.Get("database.user").(string),
			Password: config.Get("database.password").(string),
		}, ServerConfig{
			Host: config.Get("server.host").(string),
			Port: uint16(config.Get("server.port").(int64)),
		}, nil
}

func init() {
	config, configServer, err := loadConfig("config.toml")
	if err != nil {
		slog.Error("Failed to load config: ", "error", err)
		return
	}

	Conn = config
	Server = configServer
}
