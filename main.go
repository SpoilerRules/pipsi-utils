package main

import (
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/log"
	"golang.org/x/sys/windows"
	"os"
	"os/exec"
	"slices"
	"strings"
	"syscall"
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
	pauseAndExit()
}

func showMainMenu() {
	clearTerminal()
	var action string

	//goland:noinspection SqlNoDataSourceInspection
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select an action from the options below").
				Options(
					huh.NewOptions(
						"Install Pipsi", "Check for updates", "Manage Pipsi shortcuts", "Doesn't have 'Insert' key?",
					)...,
				).
				Value(&action).
				Height(6),
		),
	).WithAccessible(IsAccessible())

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
		// case: "Open Installation Folder"
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
