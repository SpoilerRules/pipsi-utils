package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
)

type GameCheat struct {
	GameTitle             string `yaml:"game_title"`
	CheatInstallationLink string `yaml:"cheat_installation_link"`
	InstallationFolder    string `yaml:"installation_folder"`
	IsDiscontinued        bool   `yaml:"is_discontinued"`
}

type Config struct {
	Cheats []GameCheat `yaml:"cheat_catalog"`
}

var config Config

func loadCheatCatalogFromURL(url string) error {
	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch configuration from URL %s: %w", url, err)
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			err = fmt.Errorf("failed to close response body: %w", closeErr)
		}
	}()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch configuration: received status %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := yaml.Unmarshal(body, &config); err != nil {
		return fmt.Errorf("failed to decode YAML from URL: %w", err)
	}

	return nil
}

func (cfg *Config) getSupportedGameTitles() []string {
	var titles []string
	for _, cheat := range cfg.Cheats {
		if !cheat.IsDiscontinued {
			titles = append(titles, cheat.GameTitle)
		}
	}
	return titles
}

func (cfg *Config) getInstallationLink(title string) string {
	for _, cheat := range cfg.Cheats {
		if cheat.GameTitle == title {
			return cheat.CheatInstallationLink
		}
	}
	fmt.Printf("Error: InstallationData title '%s' not found. The installation link could not be retrieved.\n", title)
	return ""
}

func (cfg *Config) getInstallationFolder(title string) string {
	for _, cheat := range cfg.Cheats {
		if cheat.GameTitle == title {
			return cheat.InstallationFolder
		}
	}
	fmt.Printf("Error: InstallationData title '%s' not found. The installation folder could not be retrieved.\n", title)
	return ""
}
