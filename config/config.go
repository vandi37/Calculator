package config

import (
	"encoding/json"
	"os"

	"github.com/vandi37/vanerrors"
)

// Errors
const (
	ErrorOpeningConfig  = "error opening config"
	ErrorDecodingConfig = "error decoding config"
)

// THe config
type Config struct {
	Port int    `json:"port"`
	Path string `json:"path"`
}

// Loads config
func LoadConfig(filepath string) (*Config, error) {
	file, err := os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, vanerrors.NewWrap(ErrorOpeningConfig, err, vanerrors.EmptyHandler)
	}

	defer file.Close()

	var cfg Config
	err = json.NewDecoder(file).Decode(&cfg)
	if err != nil {
		return nil, vanerrors.NewWrap(ErrorDecodingConfig, err, vanerrors.EmptyHandler)
	}

	return &cfg, nil
}
