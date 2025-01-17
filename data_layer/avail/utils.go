package avail

import (
	"os"

	"github.com/pelletier/go-toml"
)

func writeConfigToTOML(path string, c Avail) error {
	tomlBytes, err := toml.Marshal(c)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, tomlBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

func loadConfigFromTOML(path string) (Avail, error) {
	var config Avail
	tomlBytes, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = toml.Unmarshal(tomlBytes, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
