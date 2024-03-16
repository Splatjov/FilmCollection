package main

import (
	"errors"
	"github.com/jackc/pgx"
	"github.com/pelletier/go-toml"
	"os"
)

type ServerConfig struct {
	Host string
	Port uint16
}

func loadConfig(Path string) (pgx.ConnConfig, ServerConfig, error) {
	content, err := os.ReadFile(Path)
	if err != nil {
		return pgx.ConnConfig{}, ServerConfig{}, errors.New("ошибка при открытии файла конфигурации")
	}
	config, err := toml.Load(string(content))
	if err != nil {
		return pgx.ConnConfig{}, ServerConfig{}, errors.New("ошибка чтении файла конфигурации")
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
