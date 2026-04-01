package config

import (
	"fmt"
	"os"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Prometheus struct {
		Address string `yaml:"address"`
	} `yaml:"prometheus"`

	Storage struct {
		QueryLogDir  string `yaml:"query_log_dir"`
		LookbackDays int    `yaml:"lookback_days"`
	} `yaml:"storage"`

	Grafana struct {
		Address string `yaml:"address"`
		ApiKey  string `yaml:"api_key"`
	} `yaml:"grafana"`

	Output struct {
		File string `yaml:"file"`
	} `yaml:"output"`
}

func LoadConfig(path string) (*Config, error) {
	conf := &Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, conf); err != nil {
		return nil, err
	}

	// Validation: Check if log dir exists if provided
	if conf.Storage.QueryLogDir != "" {
		info, err := os.Stat(conf.Storage.QueryLogDir)
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("local query_log_dir does not exist: %s", conf.Storage.QueryLogDir)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("query_log_dir is not a directory: %s", conf.Storage.QueryLogDir)
		}
	}

	return conf, nil
}