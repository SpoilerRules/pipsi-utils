package main

import (
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/log"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type installationInfo struct {
	installationFolder string
	downloadURL        string
	gameRegion         Region
	LatestVersion      string
}

type installationProcess struct {
	games                 map[string]installationInfo
	updateMode            bool
	selectedGames         []string
	installationLinks     []string
	installationFolders   []string
	successfullyInstalled []string
}

func newInstallationProcess() *installationProcess {
	return &installationProcess{
		games: make(map[string]installationInfo),
	}
}

func (ip *installationProcess) AddGame(title string) {
	ip.games[title] = installationInfo{}
}

func (ip *installationProcess) updateGameRegion(title string, region Region) {
	if game, exists := ip.games[title]; exists {
		game.gameRegion = region
		ip.games[title] = game
	} else {
		log.Fatalf("Game title '%s' does not exist.", title)
	}
}

func (ip *installationProcess) updateGameLatestRelease(title, latestRelease string) {
	if game, exists := ip.games[title]; exists {
		game.LatestVersion = latestRelease
		ip.games[title] = game
	} else {
		log.Fatalf("Game title '%s' does not exist.", title)
	}
}

func (ip *installationProcess) getGameInfo(title string) (installationInfo, bool) {
	gameInfo, exists := ip.games[title]
	return gameInfo, exists
}

func startCheatInstallation() {
	ip := newInstallationProcess()
	supportedTitles := config.getSupportedGameTitles()

	if len(supportedTitles) == 0 {
		fmt.Println("\n\nIt seems we've hit a dead end. No supported games could be found. The Pipsi project may have quietly faded away, or perhaps you're trying to use this tool long after it was abandoned. There's a chance it's just a temporary issue with the cheat catalog, maybe a lost connection or a server that no longer answers. Whatever it is, we regret the inconvenience, and we can only hope that things turn around someday.")
		time.Sleep(10 * time.Second)
		os.Exit(0)
		return
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Which game(s) would you like to install Pipsi for?").
				Options(huh.NewOptions(supportedTitles...)...).
				Value(&ip.selectedGames),
		),
	).WithAccessible(IsAccessible())

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	if len(ip.selectedGames) == 0 {
		showMainMenu()
		return
	}

	for _, game := range ip.selectedGames {
		ip.AddGame(game)
		link := config.getInstallationLink(game)
		if link == "" {
			log.Errorf("No installation link found for game: %s", game)
			continue
		}
		ip.installationLinks = append(ip.installationLinks, link)

		folder := config.getInstallationFolder(game)
		if folder == "" {
			log.Errorf("No installation folder found for game: %s", game)
			continue
		}
		ip.installationFolders = append(ip.installationFolders, folder)
	}

	if len(ip.installationLinks) == 0 || len(ip.installationLinks) != len(ip.installationFolders) {
		log.Fatal("Error: No valid installation links or folders found, or mismatched data.")
		return
	}

	ip.startDownload()

	hasInstalledCheats := len(ip.successfullyInstalled) > 0

	showShortcutMenu(ip.successfullyInstalled)

	if hasInstalledCheats {
		log.Infof("Successfully installed Pipsi for %s.", formatGameList(ip.successfullyInstalled))
	}

	fmt.Print("\n")
	if hasInstalledCheats {
		go func() {
			createPipsiShortcuts(ip.successfullyInstalled, currentDir)
		}()
		fmt.Println(
			WarningText.Render("NOTE: ") +
				HighlightText.Render("Press Insert (or FN + Insert on laptops) in-game to toggle the cheats."),
		)
	}

	pauseAndReturnToMainMenu()
}

func startCheatUpdate(updateMap map[string]map[string]string) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Recovered from panic in startCheatUpdate: %v", r)
		}
	}()

	ip := newInstallationProcess()
	ip.updateMode = true

	for title, details := range updateMap {
		downloadURL, latestRelease := details["downloadURL"], details["latestVersion"]

		ip.selectedGames = append(ip.selectedGames, title)
		ip.AddGame(title)
		ip.installationLinks = append(ip.installationLinks, downloadURL)

		ip.updateGameLatestRelease(title, latestRelease)

		folder := config.getInstallationFolder(title)
		if folder == "" {
			log.Errorf("No installation folder found for game: %s", title)
			continue
		}
		ip.installationFolders = append(ip.installationFolders, folder)
	}

	if len(ip.installationLinks) == 0 || len(ip.installationLinks) != len(ip.installationFolders) {
		log.Error("Error: No valid installation links or folders found, or mismatched data.")
		return
	}

	ip.startDownload()

	hasInstalledCheats := len(ip.successfullyInstalled) > 0
	if hasInstalledCheats {
		log.Infof("Successfully updated Pipsi for %s.", formatGameList(ip.successfullyInstalled))
	}
	fmt.Print("\n")
	pauseAndReturnToMainMenu()
}

func (ip *installationProcess) startDownload() {
	manageDefenderExclusion()
	var wg sync.WaitGroup

	joinedTitles := formatGameList(ip.selectedGames)

	log.Infof("Beginning to download Pipsi for %s...", joinedTitles)

	for i := 0; i < len(ip.installationLinks); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			title := ip.selectedGames[i]
			downloadURL := ip.installationLinks[i]
			installationFolder := ip.installationFolders[i]

			installDir := fmt.Sprintf("Pipsi Installations/%s", installationFolder)
			if err := os.MkdirAll(installDir, os.ModePerm); err != nil {
				log.Errorf("Failed to create directory for %s: %v", title, err)
				return
			}

			filePath := fmt.Sprintf("%s/x64.rar", installDir)
			if err := ip.downloadFile(downloadURL, filePath, title); err != nil {
				log.Errorf("Download failed for %s: %v", title, err)
				return
			}

			gameData := cheatInstallationData.getGame(title)
			if gameData != nil {
				installationInfo, _ := ip.getGameInfo(title)
				gameData.Installed = true
				gameData.InstalledVersion = installationInfo.LatestVersion
				gameData.Region = installationInfo.gameRegion
				if err := saveConfigToFile("installation_data.yaml", &cheatInstallationData); err != nil {
					log.Errorf(
						"Error saving installation data for Pipsi (%s). Without saving this data, shortcut creation and update checks may not work properly. Error: %v",
						title, err,
					)
					return
				}
			}
			ip.successfullyInstalled = append(ip.successfullyInstalled, title)
		}(i)
	}
	wg.Wait()
}

func (ip *installationProcess) downloadFile(url, filePath, title string) error {
	var tagName, downloadURL string
	action := func() {
		if !ip.updateMode {
			var err error
			tagName, downloadURL, err = getLatestReleaseInfo(url)
			if err != nil {
				log.Errorf("Failed to get release info for %s: %v", title, err)
				return
			}
			ip.updateGameLatestRelease(title, tagName)
		} else {
			downloadURL = url
		}

		fileResponse, err := http.Get(downloadURL)
		if err != nil {
			log.Errorf("Failed to download file from %s: %v", downloadURL, err)
			return
		}
		defer func() {
			if closeErr := fileResponse.Body.Close(); closeErr != nil {
				log.Warnf("Error closing response body for %s: %v", title, closeErr)
			}
		}()

		if fileResponse.StatusCode != http.StatusOK {
			log.Errorf("Failed to download file for %s. HTTP status: %d", title, fileResponse.StatusCode)
			return
		}

		file, err := os.Create(filePath)
		if err != nil {
			log.Errorf("Failed to create file for %s: %v", title, err)
			return
		}
		defer func() {
			if closeErr := file.Close(); closeErr != nil {
				log.Warnf("Encountered an issue while closing the file for %s: %v", title, closeErr)
			}
		}()

		log.Debugf(
			"Downloading %s for %s (%.2f MB)...", downloadURL, title, float64(fileResponse.ContentLength)/(1024*1024),
		)

		if _, err := io.Copy(file, fileResponse.Body); err != nil {
			log.Errorf("Error downloading file for %s: %v", title, err)
			return
		}

		log.Debugf("Download complete for %s. File saved to '%s'.", title, filePath)
	}

	spinnerTitle := "Downloading Pipsi for %s..."
	if ip.updateMode {
		spinnerTitle = "Updating Pipsi for %s..."
	}
	spinnerTitle = fmt.Sprintf(spinnerTitle, formatGameList(ip.selectedGames))
	if err := spinner.New().Title(spinnerTitle).Action(action).Run(); err != nil {
		log.Errorf("Spinner widget error: %v", err)
		return err
	}

	err := ip.unzip(filePath, title)
	if err != nil {
		log.Errorf("Failed to unzip the file for Pipsi (%s): %v", title, err)
		return err
	}

	return nil
}

func (ip *installationProcess) unzip(filePath, title string) error {
	outputDir := fmt.Sprintf("Pipsi Installations/%s", config.getInstallationFolder(title))
	err := extractRar(filePath, outputDir)
	if err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	globalDir := filepath.Join(outputDir, "Global")

	if _, err := os.Stat(globalDir); err == nil {
		if err := extractFiles(globalDir, outputDir); err != nil {
			return fmt.Errorf("failed to extract contents of 'Global' folder: %w", err)
		}
		ip.updateGameRegion(title, Global)
	} else if os.IsNotExist(err) {
		ip.updateGameRegion(title, Unknown)
	} else {
		return fmt.Errorf("failed to stat 'Global' directory: %w", err)
	}

	return nil
}

func extractFiles(source, destination string) error {
	files, err := os.ReadDir(source)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	for _, file := range files {
		srcPath := filepath.Join(source, file.Name())
		destPath := filepath.Join(destination, file.Name())

		if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory for file %s: %w", file.Name(), err)
		}

		if err := copyFile(srcPath, destPath); err != nil {
			return fmt.Errorf("failed to extract file %s: %w", file.Name(), err)
		}
	}

	return nil
}

func copyFile(source, destination string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := in.Close(); closeErr != nil {
			err = fmt.Errorf("error closing input file: %w", closeErr)
		}
	}()

	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil {
			if err != nil {
				err = fmt.Errorf("error closing output file after write: %w", closeErr)
			} else {
				err = fmt.Errorf("error closing output file: %w", closeErr)
			}
		}
	}()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return err
}

func formatGameList(titles []string) string {
	switch len(titles) {
	case 1:
		return titles[0]
	case 2:
		return fmt.Sprintf("%s and %s", titles[0], titles[1])
	default:
		return fmt.Sprintf(
			"%s, %s, and %s", titles[0], strings.Join(titles[1:len(titles)-1], ", "), titles[len(titles)-1],
		)
	}
}
