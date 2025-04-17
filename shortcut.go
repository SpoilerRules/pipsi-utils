package main

import (
	"embed"
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/log"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
)

//go:embed "Game Icons/*.ico"
var embeddedIcons embed.FS

type shortcutLocation int

const (
	desktop shortcutLocation = iota
	currentDir
	startMenu
)

func showShortcutMenu(games []string) {
	var createShortcuts bool
	var selectedInstallations []string

	var prompt, mst string
	if len(games) == 1 {
		prompt = fmt.Sprintf(
			"Would you like to create a desktop shortcut for %s (Pipsi for %s)?",
			config.getInstallationFolder(games[0]), games[0],
		)
		mst = "Select the Pipsi installation for which you would like to create a desktop shortcut"
	} else if len(games) > 1 {
		prompt = fmt.Sprintf("Would you like to create desktop shortcuts for Pipsi across %d games?", len(games))
		mst = "Select the Pipsi installations for which you would like to create desktop shortcuts"
	} else {
		return
	}

	var shortcutOptions []huh.Option[string]
	action := func() {
		addedTitles := make(map[string]bool)

		for _, title := range games {
			option := fmt.Sprintf("%s", title)
			shortcutOptions = append(
				shortcutOptions,
				huh.NewOption(fmt.Sprintf("%s (%s)", config.getInstallationFolder(title), title), option).
					Selected(true),
			)
			addedTitles[title] = true
		}

		additionalInstalledTitles := cheatInstallationData.getInstalledTitles()
		for _, title := range additionalInstalledTitles {
			if !addedTitles[title] {
				option := fmt.Sprintf("%s", title)
				shortcutOptions = append(
					shortcutOptions,
					huh.NewOption(fmt.Sprintf("%s (%s)", config.getInstallationFolder(title), title), option).
						Selected(false),
				)
			}
		}
	}
	if err := spinner.New().Title("Preparing shortcut options...").Action(action).Run(); err != nil {
		log.Errorf("Spinner widget error: %v", err)
		pauseAndReturnToMainMenu()
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(prompt).
				Value(&createShortcuts).
				Affirmative("Yes").
				Negative("No"),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(mst).
				Options(shortcutOptions...).
				Value(&selectedInstallations),
		).WithHideFunc(
			func() bool {
				return !createShortcuts
			},
		),
	)

	if err := form.Run(); err != nil {
		log.Fatal(err)
	}

	if !createShortcuts {
		return
	}

	if len(selectedInstallations) < 1 {
		log.Infof("No installations were selected to create shortcuts for %s. Skipping shortcut creation.", mst)
		return
	}

	if scs := createPipsiShortcuts(selectedInstallations, desktop); len(scs) > 0 {
		log.Infof(
			"Shortcut(s) created successfully for %s.",
			formatGameList(scs),
		)
	} else {
		log.Infof("No shortcuts were created for %s.", formatGameList(selectedInstallations))
	}
}

func showAdvancedShortcutMenu() {
	installedTitles := cheatInstallationData.getInstalledTitles()
	if len(installedTitles) == 0 {
		fmt.Print(
			HighlightText.Render("No Pipsi installations detected.\n\n") +
				"Please return to the main menu and select " +
				BoldCyan.Render("\"Install Pipsi\"") +
				" to get started.\n\n",
		)
		pauseAndReturnToMainMenu()
		return
	}

	/*	fmt.Println(
		NoticePrefix.Render("NOTE: ") +
			HighlightText.Render("If the Pipsi installation you're looking for isn't listed here, ") +
			HighlightText.Render("ensure it is installed first.\n") +
			WarningText.Render("Quick return: Press Enter three times without selections to go back.\n"),
	)*/

	type optionData struct {
		display string
		value   string
	}
	var optionsData []optionData

	action := func() {
		for _, title := range installedTitles {
			display := fmt.Sprintf("%s (%s)", config.getInstallationFolder(title), title)
			optionsData = append(
				optionsData, optionData{
					display: display,
					value:   title,
				},
			)
		}
	}

	if err := spinner.New().Title("Preparing shortcut options...").Action(action).Run(); err != nil {
		log.Errorf("Spinner error: %v", err)
		pauseAndReturnToMainMenu()
		return
	}

	generateOptions := func(data []optionData) []huh.Option[string] {
		options := make([]huh.Option[string], len(data))
		for i, d := range data {
			options[i] = huh.NewOption(d.display, d.value)
		}
		return options
	}

	var (
		desktopSelections, dirSelections, menuSelections []string
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select installations for desktop shortcuts").
				Options(generateOptions(optionsData)...).
				Value(&desktopSelections),

			huh.NewMultiSelect[string]().
				Title("Create shortcuts in current directory").
				Description("Makes shortcuts adjacent to this executable").
				Options(generateOptions(optionsData)...).
				Value(&dirSelections),

			huh.NewMultiSelect[string]().
				Title("Start Menu shortcuts").
				Description("Creates in Windows Start Menu Programs folder").
				Options(generateOptions(optionsData)...).
				Value(&menuSelections),
		),
	)

	if err := form.Run(); err != nil {
		log.Fatal(err)
	}

	totalSelections := len(desktopSelections) + len(dirSelections) + len(menuSelections)
	if totalSelections == 0 {
		returnToMainMenu()
		return
	}

	var createdShortcuts []string
	createAction := func() {
		createdShortcuts = append(
			createdShortcuts,
			createPipsiShortcuts(desktopSelections, desktop)...,
		)
		createdShortcuts = append(
			createdShortcuts,
			createPipsiShortcuts(dirSelections, currentDir)...,
		)
		createdShortcuts = append(
			createdShortcuts,
			createPipsiShortcuts(menuSelections, startMenu)...,
		)
	}

	if err := spinner.New().Title("Creating shortcuts...").Action(createAction).Run(); err != nil {
		log.Errorf("Creation failed: %v", err)
	}

	if len(createdShortcuts) > 0 {
		log.Infof(
			"Successfully created shortcuts for: %s",
			formatGameList(createdShortcuts),
		)
	}
	pauseAndReturnToMainMenu()
}

func createPipsiShortcuts(selectedInstallations []string, location shortcutLocation) []string {
	err := extractGameIcons()
	if err != nil {
		log.Errorf("Failed to prepare icons for shortcuts: %v", err)
		return nil
	}

	var scs []string
	for _, game := range selectedInstallations {
		exePath, err := os.Executable()
		if err != nil {
			log.Errorf("Failed to get executable path: %v", err)
			pauseAndReturnToMainMenu()
			return nil
		}

		exeDir := filepath.Dir(exePath)
		cif := config.getInstallationFolder(game)
		installationFolderPath := filepath.Join(exeDir, cif)
		appPath := filepath.Join(filepath.Join(exeDir, "Pipsi Installations", cif), "Launcher.exe")
		iconPath := filepath.Join(os.TempDir(), "Game Icons", fmt.Sprintf("%s.ico", cif))

		if err = createShortcut(cif, appPath, iconPath, installationFolderPath, location); err != nil {
			log.Errorf("Failed to create shortcut for %s: %v", game, err)
		} else {
			scs = append(scs, game)
		}
	}
	return scs
}

func createShortcut(shortcutName, appPath, iconPath, workingDir string, location shortcutLocation) error {
	shortcutName = regexp.MustCompile(`[<>:"/\\|?*]`).ReplaceAllString(shortcutName, "")

	var shortcutPath string
	switch location {
	case desktop:
		shortcutPath = filepath.Join(filepath.Join(os.Getenv("USERPROFILE"), "Desktop"), shortcutName+".lnk")
	case currentDir:
		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("error getting executable path: %v", err)
		}
		shortcutPath = filepath.Join(filepath.Dir(exePath), shortcutName+".lnk")
	case startMenu:
		shortcutPath = filepath.Join(
			filepath.Join(
				os.Getenv("APPDATA"),
				"Microsoft", "Windows", "Start Menu", "Programs",
			), shortcutName+".lnk",
		)
	default:
		return fmt.Errorf("invalid shortcut location: %v", location)
	}

	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		return fmt.Errorf("error initializing OLE: %v", err)
	}
	defer ole.CoUninitialize()

	shellLink, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		return fmt.Errorf("error creating WScript.Shell object: %v", err)
	}
	defer shellLink.Release()

	wshell, err := shellLink.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Errorf("error querying IDispatch: %v", err)
	}
	defer wshell.Release()

	cs, err := oleutil.CallMethod(wshell, "CreateShortcut", shortcutPath)
	if err != nil {
		return fmt.Errorf("error creating shortcut: %v", err)
	}
	shortcut := cs.ToIDispatch()
	defer shortcut.Release()

	if _, err := oleutil.PutProperty(shortcut, "TargetPath", appPath); err != nil {
		return fmt.Errorf("error setting TargetPath: %v", err)
	}
	if _, err := oleutil.PutProperty(shortcut, "IconLocation", iconPath); err != nil {
		return fmt.Errorf("error setting IconLocation: %v", err)
	}
	if _, err := oleutil.PutProperty(shortcut, "WorkingDirectory", workingDir); err != nil {
		return fmt.Errorf("error setting WorkingDirectory: %v", err)
	}
	if _, err := oleutil.CallMethod(shortcut, "Save"); err != nil {
		return fmt.Errorf("error saving shortcut: %v", err)
	}

	return nil
}

func extractGameIcons() error {
	targetDir := filepath.Join(os.TempDir(), "Game Icons")
	extractionNeeded := false

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		extractionNeeded = true
	} else {
		entries, err := embeddedIcons.ReadDir("Game Icons")
		if err != nil {
			log.Warn("Corrupted icon manifest, regenerating icons")
			extractionNeeded = true
		} else {
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				targetPath := filepath.Join(targetDir, entry.Name())
				if _, err := os.Stat(targetPath); os.IsNotExist(err) {
					extractionNeeded = true
					break
				}
			}
		}
	}

	if !extractionNeeded {
		log.Debug("Icons already cached")
		return nil
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		log.Error("Failed to create icon cache directory", "path", targetDir, "error", err)
		return err
	}

	entries, err := embeddedIcons.ReadDir("Game Icons")
	if err != nil {
		log.Error("Failed to read embedded icons", "error", err)
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		data, err := embeddedIcons.ReadFile(path.Join("Game Icons", entry.Name()))
		if err != nil {
			log.Error("Failed to read embedded icon", "file", entry.Name(), "error", err)
			return err
		}

		targetPath := filepath.Join(targetDir, entry.Name())
		if err := os.WriteFile(targetPath, data, 0644); err != nil {
			log.Error("Failed to write icon cache", "path", targetPath, "error", err)
			return err
		}
	}

	log.Debug("Icon cache updated")
	return nil
}
