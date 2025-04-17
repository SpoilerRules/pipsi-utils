package main

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/log"
	"golang.org/x/sys/windows"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"unsafe"
)

var (
	shell32      = syscall.NewLazyDLL("shell32.dll")
	shellExecute = shell32.NewProc("ShellExecuteW")
)

var localMode bool

func main() {
	log.SetReportTimestamp(false)
	log.SetReportCaller(false)
	if !windows.GetCurrentProcessToken().IsElevated() {
		elevate()
		os.Exit(0)
	}

	action := func() {
		cmd := exec.Command("cmd", "/C", "title", "Pipsi Utilities")
		if err := cmd.Run(); err != nil {
			log.Warnf("Failed to set terminal title: %v", err)
		}

		localMode = slices.Contains(os.Args[1:], "--local")
		if localMode {
			if err := loadYAMLFromFile("cheat_catalog.yaml", &config); err != nil {
				log.Fatalf("Failed to load local cheat catalog: %v", err)
			}
		} else {
			if err := loadCheatCatalogFromURL("https://raw.githubusercontent.com/SpoilerRules/pipsi-utils-cloud/refs/heads/main/cheat_catalog.yaml"); err != nil {
				log.Fatalf("Failed to fetch cheat catalog from the remote server: %v", err)
			}
		}

		if err := loadYAMLFromFile("installation_data.yaml", &cheatInstallationData); err != nil {
			log.Infof("Failed to load installation data file: %v. Attempting to download a new one...", err)
			if err := downloadYAMLFromURL(
				"https://raw.githubusercontent.com/SpoilerRules/pipsi-utils-cloud/refs/heads/main/installation_data.yaml",
				"installation_data.yaml",
			); err != nil {
				log.Fatalf("Failed to download installation data file: %v", err)
			}

			if err := loadYAMLFromFile("installation_data.yaml", &cheatInstallationData); err != nil {
				log.Fatalf("Failed to load downloaded installation data file: %v", err)
			}
		}
	}
	if err := spinner.New().Title("Initializing resources...").Action(action).Run(); err != nil {
		log.Errorf("Spinner widget error: %v", err)
		pauseAndReturnToMainMenu()
	}

	showMainMenu()
}

func showMainMenu() {
	clearTerminal()
	var action string

	//goland:noinspection SqlNoDataSourceInspection
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select an action from the options below").
				OptionsFunc(
					func() []huh.Option[string] {
						options := []string{
							"Install Pipsi",
							"Check for updates",
							"Manage Pipsi shortcuts",
							"Open installation folder",
							"Doesn't have 'Insert' key?",
						}

						if len(cheatInstallationData.getInstalledTitles()) == 0 {
							options = slices.DeleteFunc(
								options, func(s string) bool {
									return s == "Check for updates" || s == "Open Installation Folder"
								},
							)
						}

						return huh.NewOptions(options...)
					}, &cheatInstallationData,
				).
				Value(&action).
				Height(7),
		),
	)

	if err := form.Run(); err != nil {
		log.Fatal(err)
	}

	switch action {
	case "Install Pipsi":
		startCheatInstallation()
	case "Check for updates":
		showUpdateMenu()
	case "Manage Pipsi shortcuts":
		showAdvancedShortcutMenu()
		// case: "Pipsi Unpacker":
	case "Doesn't have 'Insert' key?":
		clearTerminal()
		err := exec.Command("cmd", "/c", "osk").Start()
		if err != nil {
			log.Errorf("Unable to launch on-screen keyboard: %v", err)
			pauseAndReturnToMainMenu()
		}
		log.Infof("On-screen keyboard launched. Use it to access the Insert key.")
		pauseAndReturnToMainMenu()
	case "Open Installation Folder":
		workingDir, err := os.Getwd()
		if err != nil {
			log.Errorf("Failed to get working directory: %v", err)
			pauseAndReturnToMainMenu()
		}

		installPath := filepath.Join(workingDir, "Pipsi Installations")

		dirInfo, err := os.Stat(installPath)
		switch {
		case os.IsNotExist(err):
			log.Errorf("Installation directory not found: %s", installPath)
			pauseAndReturnToMainMenu()
		case err != nil:
			log.Errorf("Failed to access installation directory: %v", err)
			pauseAndReturnToMainMenu()
		case !dirInfo.IsDir():
			log.Errorf("Invalid installation path (not a directory): %s", installPath)
			pauseAndReturnToMainMenu()
		}

		const (
			openVerb    = "open"
			swShow      = 5
			successCode = 32
		)

		pathPtr, err := syscall.UTF16PtrFromString(installPath)
		if err != nil {
			log.Errorf("Failed to encode path string: %v", err)
			pauseAndReturnToMainMenu()
		}

		verbPtr, err := syscall.UTF16PtrFromString(openVerb)
		if err != nil {
			log.Errorf("Failed to encode action verb: %v", err)
			pauseAndReturnToMainMenu()
		}

		ret, _, err := shellExecute.Call(
			0,
			uintptr(unsafe.Pointer(verbPtr)),
			uintptr(unsafe.Pointer(pathPtr)),
			0,
			0,
			swShow,
		)

		switch {
		case !errors.Is(err, syscall.Errno(0)):
			log.Errorf("Directory open operation failed: %v", err)
		case ret <= successCode:
			log.Errorf("Directory open failed with system code %d", ret)
		default:
			log.Info("Successfully opened installation directory at %s", installPath)
		}
	default:
		fmt.Println("Unknown action. Exiting.")
	}

}

// https://gist.github.com/jerblack/d0eb182cc5a1c1d92d92a4c4fcc416c6
func elevate() {
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := strings.Join(os.Args[1:], " ")

	verbPtr, _ := syscall.UTF16PtrFromString("runas")
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argPtr, _ := syscall.UTF16PtrFromString(args)

	if err := windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, 1); err != nil {
		log.Fatalf("Failed to elevate process: %v", err)
	}
}
