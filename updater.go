package main

import (
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/log"
)

func showUpdateMenu() {
	foundUpdate := false
	updates := make(map[string]map[string]string)

	action := func() {
		installedVersions := cheatInstallationData.getInstalledVersions()
		updates = checkForUpdates(installedVersions)

		foundUpdate = len(updates) > 0
	}

	if err := spinner.New().Title("Checking for updates...").Action(action).Run(); err != nil {
		log.Errorf("Spinner widget error: %v", err)
		pauseAndReturnToMainMenu()
		return
	}

	if !foundUpdate {
		fmt.Println(StatusText.Render("No updates available, all Pipsi installations are up to date."))
		pauseAndReturnToMainMenu()
		return
	}

	var updateOptions []huh.Option[string]
	action = func() {
		for gameTitle, details := range updates {
			installedVersion := cheatInstallationData.getInstalledVersion(gameTitle)
			latestVersion := details["latestVersion"]

			optionText := fmt.Sprintf(
				"%s (%s %s %s)",
				gameTitle,
				installedVersion,
				"→",
				latestVersion,
			)

			updateOptions = append(updateOptions, huh.NewOption(optionText, gameTitle).Selected(true))
		}
	}

	if err := spinner.New().Title("Preparing options...").Action(action).Run(); err != nil {
		log.Errorf("Spinner widget error: %v", err)
		pauseAndReturnToMainMenu()
		return
	}

	var selectedInstallations []string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select the Pipsi installation(s) for update").
				Options(updateOptions...).
				Value(&selectedInstallations),
		),
	)

	if err := form.Run(); err != nil {
		log.Fatal(err)
	}

	updateMap := make(map[string]map[string]string)

	action = func() {
		for _, selectedTitle := range selectedInstallations {
			if details, ok := updates[selectedTitle]; ok {
				updateMap[selectedTitle] = map[string]string{
					"downloadURL":   details["downloadURL"],
					"latestVersion": details["latestVersion"],
				}
			}
		}
	}

	if err := spinner.New().Title("Preparing to start the update(s)...").Action(action).Run(); err != nil {
		log.Errorf("Spinner widget error: %v", err)
		pauseAndReturnToMainMenu()
		return
	}

	startCheatUpdate(updateMap)
}

func checkForUpdates(installedVersions map[string]string) map[string]map[string]string {
	updates := make(map[string]map[string]string)

	for gameTitle, installedVersion := range installedVersions {
		repo := config.getInstallationLink(gameTitle)

		latestVersion, downloadURL, err := getLatestReleaseInfo(repo)
		if err != nil {
			log.Debugf(
				"Error checking for updates for %s, skipping installation for %s: %v\n", gameTitle,
				config.getInstallationFolder(gameTitle), err,
			)
			continue
		}

		if installedVersion != latestVersion {
			log.Debugf(
				"Update available for %s: installed version %s, Latest version %s\n", gameTitle, installedVersion,
				latestVersion,
			)

			updates[gameTitle] = map[string]string{
				"latestVersion": latestVersion,
				"downloadURL":   downloadURL,
			}
		} else {
			log.Debugf(
				"No update available for %s: installed version is up to date (%s)\n", gameTitle, installedVersion,
			)
		}
	}

	return updates
}
