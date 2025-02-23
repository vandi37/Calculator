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

type Config struct {
	Port                 Port `json:"port"`
	Path                 Path `json:"path"`
	TimeAdditionMs       int  `json:"time_addition_ms"`
	TimeSubtractionMs    int  `json:"time_subtraction_ms"`
	TimeMultiplicationMs int  `json:"time_multiplication_ms"`
	TimeDivisionMs       int  `json:"time_division_ms"`
	ComputingPower       int  `json:"computing_power"`
	MaxAgentErrors       int  `json:"max_agent_errors"`
	AgentPeriodicityMs   int  `json:"agent_periodicity"`
}

type Port struct {
	Api int `json:"api"`
}

type Path struct {
	Calc string `json:"calc"`
	Get  string `json:"get"`
	Task string `json:"task"`
}

func LoadConfig(filepath string) (*Config, error) {
	file, err := os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, vanerrors.Wrap(ErrorOpeningConfig, err)
	}

	defer file.Close()

	var cfg Config
	err = json.NewDecoder(file).Decode(&cfg)
	if err != nil {
		return nil, vanerrors.Wrap(ErrorDecodingConfig, err)
	}

	return &cfg, nil
}
