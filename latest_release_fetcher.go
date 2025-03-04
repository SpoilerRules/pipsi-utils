package main

import (
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/log"
	"net/http"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// getLatestReleaseInfo fetches the latest release version (tag_name) and the first asset's download URL from the GitHub API
func getLatestReleaseInfo(repo string) (string, string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

	response, err := http.Get(url)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			log.Warnf("Error closing response body: %v", closeErr)
		}
	}()

	if response.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("failed to fetch latest release, status code: %d", response.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(response.Body).Decode(&release); err != nil {
		return "", "", fmt.Errorf("failed to parse release JSON: %w", err)
	}

	if len(release.Assets) == 0 {
		return release.TagName, "", fmt.Errorf("no assets found in the latest release")
	}

	return release.TagName, release.Assets[0].BrowserDownloadURL, nil
}
