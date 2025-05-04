package config

import (
	"github.com/goloop/env"
	"github.com/vandi37/vanerrors"
)

// Errors
const (
	ErrorDecodingConfig = "error decoding config"
)

type Config struct {
	Port              int    `env:"PORT" def:"8080"`
	GRPCProt          int    `env:"GRPC_PORT" def:"50550"`
	Time              Time   `env:"TIME"`
	MongoUri          string `env:"MONGO_URI"`
	ResetTaskDuration string `env:"RESET_TASK_DURATION" def:"1m"`
	JWT               JWT    `env:"JWT"`
	LogFile           string `env:"LOG_FILE" def:"logs.log"`
}

type JWT struct {
	Secret    string `env:"SECRET"`
	Expires   string `env:"EXP" def:"24h"`
	NotBefore string `env:"NBF" def:"1ms"` // Creating little delay for safety
}

type Time struct {
	AdditionMs       int32 `env:"ADDITION_MS" def:"10"`
	SubtractionMs    int32 `env:"SUBTRACTION_MS" def:"10"`
	MultiplicationMs int32 `env:"MULTIPLICATION_MS" def:"10"`
	DivisionMs       int32 `env:"DIVISION_MS" def:"10"`
}

func LoadConfig() (*Config, error) {
	var cfg Config

	if err := env.Unmarshal("", &cfg); err != nil {
		return nil, vanerrors.Wrap(ErrorDecodingConfig, err)
	}

	return &cfg, nil
}
