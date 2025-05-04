package config

import (
	"github.com/goloop/env"
)

type Config struct {
	GrpcPath       string `env:"GRPC_PATH" def:"localhost:50051"`
	ComputingPower int    `env:"COMPUTING_POWER" def:"10"`
	RetryCount     int    `env:"RETRY_COUNT" def:"5"`
	LogFile        string `env:"LOG_FILE" def:"logs.log"`

}

func LoadConfig() (*Config, error) {
	var cfg Config

	if err := env.Unmarshal("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
