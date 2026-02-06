package config

import (
	"encoding/json"
	"os"
)

type Settings struct {
	APIURL string `json:"api_url"`
}

func Load(path string) (Settings, error) {
	var cfg Settings
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Settings{}, err
	}
	return cfg, nil
}

func Save(path string, cfg Settings) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}
