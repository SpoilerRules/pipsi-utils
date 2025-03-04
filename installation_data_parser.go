package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
)

type Region string

const (
	Global  Region = "Global"
	Sea     Region = "Sea"
	China   Region = "China"
	Unknown Region = "Unknown"
	None    Region = "None"
)

type InstallationData struct {
	GameTitle        string `yaml:"title"`
	Installed        bool   `yaml:"installed"`
	InstalledVersion string `yaml:"installed_version"`
	Region           Region `yaml:"region"`
}

type Games struct {
	GamesList []InstallationData `yaml:"installation_data"`
}

var cheatInstallationData = Games{
	GamesList: []InstallationData{},
}

func loadYAMLFromFile(filename string, out interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", filename, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = fmt.Errorf("failed to close %s: %w", filename, closeErr)
		}
	}()

	if err := yaml.NewDecoder(file).Decode(out); err != nil {
		return fmt.Errorf("failed to decode %s: %w", filename, err)
	}

	return nil
}

func downloadYAMLFromURL(url, filename string) error {
	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", url, err)
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			err = fmt.Errorf("failed to close response body for %s: %w", url, closeErr)
		}
	}()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s: server returned %s", url, response.Status)
	}

	out, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", filename, err)
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil {
			err = fmt.Errorf("failed to close %s: %w", filename, closeErr)
		}
	}()

	if _, err := io.Copy(out, response.Body); err != nil {
		return fmt.Errorf("failed to write to %s: %w", filename, err)
	}

	return nil
}

func saveConfigToFile(filename string, in interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", filename, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = fmt.Errorf("failed to close %s: %w", filename, closeErr)
		}
	}()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	if err := encoder.Encode(in); err != nil {
		return fmt.Errorf("failed to encode %s: %w", filename, err)
	}

	return nil
}

// Returns a map of installed game titles with their installed versions
func (g *Games) getInstalledVersions() map[string]string {
	installedVersions := make(map[string]string)
	for _, game := range g.GamesList {
		if game.Installed {
			installedVersions[game.GameTitle] = game.InstalledVersion
		}
	}
	return installedVersions
}

func (g *Games) getInstalledVersion(title string) string {
	for _, game := range g.GamesList {
		if game.GameTitle == title {
			return game.InstalledVersion
		}
	}
	return ""
}

func (g *Games) updateInstalledVersion(title, version string) bool {
	for i, game := range g.GamesList {
		if game.GameTitle == title {
			g.GamesList[i].InstalledVersion = version
			return true
		}
	}
	return false
}

func (g *Games) setContinent(title string, continent Region) bool {
	for i, game := range g.GamesList {
		if game.GameTitle == title {
			g.GamesList[i].Region = continent
			return true
		}
	}
	return false
}

func (g *Games) getGame(title string) *InstallationData {
	for i, game := range g.GamesList {
		if game.GameTitle == title {
			return &g.GamesList[i]
		}
	}
	return nil
}

func (g *Games) getInstalledTitles() []string {
	var titles []string
	for _, game := range g.GamesList {
		if game.Installed {
			titles = append(titles, game.GameTitle)
		}
	}
	return titles
}
