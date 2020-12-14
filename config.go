package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type config struct {
	Slack struct {
		ClientID     string `toml:"ClientID,omitempty"`
		ClientSecret string `toml:"ClientSecret,omitempty"`
		AccessToken  string `toml:"AccessToken,omitempty"`
	}
}

const configFileName string = "slack-file.conf"

func determineConfigPath(defaultValue string) string {
	if defaultValue != "" {
		return defaultValue
	}

	// wd
	wdConfigPath := filepath.Join(".", configFileName)
	if _, err := os.Stat(wdConfigPath); err == nil {
		return wdConfigPath
	}

	// exe
	if exepath, err := os.Executable(); err == nil {
		exeConfigPath := filepath.Join(filepath.Dir(exepath), configFileName)
		if _, err := os.Stat(exeConfigPath); err == nil {
			return exeConfigPath
		}
	}

	return wdConfigPath
}

func loadConfig(filePath string) (*config, error) {
	filePath = determineConfigPath(filePath)

	config := &config{}
	_, err := toml.DecodeFile(filePath, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "missing %v. -> creating with minimal contents...", filePath)
		if err := saveConfig(config, filePath); err != nil {
			return config, fmt.Errorf("failed to access to config: %v", err)
		}
		fmt.Fprintf(os.Stderr, "created.\n")
	}

	return config, nil
}

func saveConfig(config *config, filePath string) error {
	filePath = determineConfigPath(filePath)

	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(config); err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, buf.Bytes(), 0700)
}
