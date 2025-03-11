package main

import (
	"errors"
	"github.com/charmbracelet/log"
	"os"
	"path/filepath"
)

const accessibleFlagFileName = "PIPSI_UTILS_ACCESSIBLE_MODE"

func IsAccessible() bool {
	_, err := os.Stat(filepath.Join(os.TempDir(), accessibleFlagFileName))
	return !os.IsNotExist(err)
}

func EnableAccessibleMode() {
	flagPath := filepath.Join(os.TempDir(), accessibleFlagFileName)

	file, err := os.OpenFile(flagPath, os.O_CREATE|os.O_RDONLY, 0444)
	if err != nil {
		log.Errorf("Failed to create access flag: %v", err)
		pauseAndReturnToMainMenu()
		return
	}
	if closeErr := file.Close(); closeErr != nil {
		log.Warnf("Failed to close the file %s: %v", flagPath, closeErr)
	}

	returnToMainMenu()
}

func DisableAccessibleMode() {
	flagPath := filepath.Join(os.TempDir(), accessibleFlagFileName)

	err := os.Remove(flagPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Errorf("Failed to remove access flag: %v", err)
		pauseAndReturnToMainMenu()
		return
	}

	returnToMainMenu()
}
