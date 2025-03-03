package config

import (
	"encoding/json"
	"os"

	"github.com/goloop/env"
	"github.com/vandi37/flags"
	"github.com/vandi37/vanerrors"
)

const (
	STD_JSON = "configs/config.json"
	STD_ENV  = ".env"
)

// Errors
const (
	ErrorOpeningConfig  = "error opening config"
	ErrorDecodingConfig = "error decoding config"
	ErrorEncodingConfig = "error encoding config"
	CantUseTwoOptions   = "cant use two options"
)

type Config struct {
	Port               int  `json:"port" env:"PORT" def:"3701"`
	Path               Path `json:"path" env:"PATH"`
	Time               Time `json:"time" env:"TIME"`
	ComputingPower     int  `json:"computing_power" env:"COMPUTING_POWER" def:"5"`
	MaxAgentErrors     int  `json:"max_agent_errors" env:"MAX_AGENT_ERRORS" def:"10"`
	AgentPeriodicityMs int  `json:"agent_periodicity" env:"AGENT_PERIODICITY" def:"100"`
}

type Time struct {
	AdditionMs       int `json:"addition_ms" env:"ADDITION_MS" def:"1000"`
	SubtractionMs    int `json:"subtraction_ms" env:"SUBTRACTION_MS" def:"2000"`
	MultiplicationMs int `json:"multiplication_ms" env:"MULTIPLICATION_MS" def:"4000"`
	DivisionMs       int `json:"division_ms" env:"DIVISION_MS" def:"8000"`
}

type Path struct {
	Calc string `json:"calc" env:"CALC" def:"/api/v1/calculate"`
	Get  string `json:"get" env:"GET" def:"/api/v1/expressions"`
	Task string `json:"task" env:"TASK" def:"/internal/task"`
}

func LoadJsonConfig(filepath string) (*Config, error) {
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0666)
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

func LoadConfigEnv(filepath string) (*Config, error) {
	if err := env.Load(filepath); err != nil {
		return nil, vanerrors.Wrap(ErrorOpeningConfig, err)
	}
	var cfg Config

	if err := env.Unmarshal("", &cfg); err != nil {
		return nil, vanerrors.Wrap(ErrorDecodingConfig, err)
	}

	return &cfg, nil
}

func (c *Config) Save(filepath string) error {
	file, err := os.OpenFile(filepath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return vanerrors.Wrap(ErrorOpeningConfig, err)
	}

	defer file.Close()

	json := json.NewEncoder(file)
	json.SetIndent("", "	")
	if err := json.Encode(*c); err != nil {
		return vanerrors.Wrap(ErrorEncodingConfig, err)
	}

	return nil
}

type LoadType struct {
	DoEnv  bool `flag:"env"`
	DoJson bool `flag:"json"`
}

var StandardConfig Config = Config{
	Port: 3701,
	Path: Path{
		Calc: "/api/v1/calculate",
		Get:  "/api/v1/expressions",
		Task: "/internal/task",
	},
	Time: Time{
		AdditionMs:       1000,
		SubtractionMs:    2000,
		MultiplicationMs: 4000,
		DivisionMs:       8000,
	},
	ComputingPower:     5,
	MaxAgentErrors:     10,
	AgentPeriodicityMs: 100,
}

func LoadConfig() (*Config, error) {
	var loadType LoadType
	flags.Args(&loadType)

	if loadType.DoJson && loadType.DoEnv {
		return nil, vanerrors.New(CantUseTwoOptions, "can't use both env and json")
	}

	if loadType.DoJson {
		return LoadJsonConfig(STD_JSON)
	} else if loadType.DoEnv {
		return LoadConfigEnv(STD_ENV)
	}

	return &StandardConfig, nil
}
