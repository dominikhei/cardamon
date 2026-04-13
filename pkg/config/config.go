// Copyright 2026 dominikhei
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the mapping of the config.yaml used to supply parameters to cardamon.
type Config struct {
	Prometheus struct {
		Address    string `yaml:"address"`
		PathPrefix string `yaml:"path_prefix"`
		Token      string `yaml:"token"`
		Username   string `yaml:"username"`
		Password   string `yaml:"password"`
	} `yaml:"prometheus"`

	Storage struct {
		QueryLogDir  string `yaml:"query_log_dir"`
		LookbackDays int    `yaml:"lookback_days"`
	} `yaml:"storage"`

	Grafana struct {
		Address    string `yaml:"address"`
		PathPrefix string `yaml:"path_prefix"`
		Token      string `yaml:"token"`
		Username   string `yaml:"username"`
		Password   string `yaml:"password"`
	} `yaml:"grafana"`

	Audit struct {
		ExcludePrefixes []string `yaml:"exclude_prefixes"`
	} `yaml:"audit"`

	Dashboard struct {
		Port int `yaml:"port"`
	} `yaml:"dashboard"`
}

// Function to load the config file.
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
