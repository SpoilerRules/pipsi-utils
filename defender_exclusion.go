package main

import (
	"bytes"
	"fmt"
	"github.com/charmbracelet/log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func manageDefenderExclusion() {
	if !isDefenderActive() {
		log.Warnf("Windows Defender is not active.\nExternal antivirus software may flag or delete Pipsi installations.\nWe are proceeding to add exclusions, but please ensure your antivirus software allows Pipsi by setting the necessary exclusions manually.")
	}

	currentDir, err := os.Getwd()
	if err != nil {
		log.Errorf("Unable to get current directory: %v", err)
		return
	}
	pipsiFolderPath := filepath.Join(currentDir, "Pipsi Installations")

	if _, err := os.Stat(pipsiFolderPath); os.IsNotExist(err) {
		if err := os.Mkdir(pipsiFolderPath, 0755); err != nil {
			log.Errorf("Unable to create Pipsi Installations folder: %v", err)
			return
		}
	}

	checkCmd := exec.Command(
		"powershell", "-Command", fmt.Sprintf(
			`$exclusions = Get-MpPreference | Select -ExpandProperty ExclusionPath; $exclusions -contains "%s"`,
			pipsiFolderPath,
		),
	)
	var checkOutput bytes.Buffer
	var checkStderr bytes.Buffer
	checkCmd.Stdout = &checkOutput
	checkCmd.Stderr = &checkStderr

	if err := checkCmd.Run(); err != nil {
		log.Errorf("Failed to check Windows Defender exclusions.\nError: %v", checkStderr.String())
		return
	}

	if strings.TrimSpace(checkOutput.String()) == "True" {
		log.Debugf("'%s' is already in Windows Defender exclusions.", pipsiFolderPath)
		return
	}

	log.Debugf("Attempting to add '%s' to Windows Defender exclusions...", pipsiFolderPath)

	addCmd := exec.Command(
		"powershell", "-Command", fmt.Sprintf(`Add-MpPreference -ExclusionPath "%s"`, pipsiFolderPath),
	)
	var addStderr bytes.Buffer
	addCmd.Stderr = &addStderr

	if err := addCmd.Run(); err != nil {
		log.Errorf(
			"Failed to add folder to Windows Defender exclusions.\nConsider disabling real-time protection or manually adding the folder.\nError: %v",
			addStderr.String(),
		)
		return
	}

	log.Infof("Successfully added '%s' to Windows Defender exclusions.", pipsiFolderPath)
}

func isDefenderActive() bool {
	cmd := exec.Command(
		"powershell", "-Command", "Get-MpPreference | Select-Object -ExpandProperty DisableRealtimeMonitoring",
	)
	var output bytes.Buffer
	cmd.Stdout = &output

	if err := cmd.Run(); err != nil {
		log.Warnf("Error checking Windows Defender status: %v\n", err)
		pauseAndExit()
	}

	trimmedOutput := strings.TrimSpace(output.String())

	return trimmedOutput == "False"
}
