package main

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/log"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

const (
	stdInputHandle       = uint32(-10 & 0xFFFFFFFF)
	enableLineInput      = 0x0002
	enableEchoInput      = 0x0004
	enableProcessedInput = 0x0001
	keyEvent             = 0x0001
)

var (
	modkernel32          = syscall.NewLazyDLL("kernel32.dll")
	procGetStdHandle     = modkernel32.NewProc("GetStdHandle")
	procGetConsoleMode   = modkernel32.NewProc("GetConsoleMode")
	procSetConsoleMode   = modkernel32.NewProc("SetConsoleMode")
	procReadConsoleInput = modkernel32.NewProc("ReadConsoleInputW")
)

type keyEventRecord struct {
	bKeyDown          int32
	wRepeatCount      uint16
	wVirtualKeyCode   uint16
	wVirtualScanCode  uint16
	unicodeChar       uint16
	dwControlKeyState uint32
}

type inputRecord struct {
	eventType uint16
	_         [2]byte
	event     [16]byte
}

func getConsoleMode(handle syscall.Handle) (uint32, error) {
	var mode uint32
	success, _, err := procGetConsoleMode.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&mode)),
	)
	if success == 0 {
		return 0, fmt.Errorf("get console mode: %w", err)
	}
	return mode, nil
}

func setConsoleMode(handle syscall.Handle, mode uint32) error {
	result, _, err := procSetConsoleMode.Call(
		uintptr(handle),
		uintptr(mode),
	)
	if result == 0 {
		return fmt.Errorf("set console mode: %w", err)
	}
	return nil
}

func waitForKeyPress() (err error) {
	handlePtr, _, callErr := procGetStdHandle.Call(uintptr(stdInputHandle))
	if handlePtr == uintptr(syscall.InvalidHandle) {
		return fmt.Errorf("failed to get console handle: %w", callErr)
	}
	consoleHandle := syscall.Handle(handlePtr)

	originalMode, err := getConsoleMode(consoleHandle)
	if err != nil {
		return fmt.Errorf("get console mode: %w", err)
	}
	defer func() {
		if restoreErr := setConsoleMode(consoleHandle, originalMode); restoreErr != nil {
			err = errors.Join(err, fmt.Errorf("restore console mode: %w", restoreErr))
		}
	}()

	rawMode := originalMode &^ (enableLineInput | enableEchoInput | enableProcessedInput)
	if err := setConsoleMode(consoleHandle, rawMode); err != nil {
		return fmt.Errorf("set raw input mode: %w", err)
	}

	var event inputRecord
	for {
		var eventsRead uint32
		success, _, callErr := procReadConsoleInput.Call(
			uintptr(consoleHandle),
			uintptr(unsafe.Pointer(&event)),
			1,
			uintptr(unsafe.Pointer(&eventsRead)),
		)
		if success == 0 {
			return fmt.Errorf("console input read failed: %w", callErr)
		}

		if event.eventType == keyEvent {
			kevt := (*keyEventRecord)(unsafe.Pointer(&event.event[0]))
			if kevt.bKeyDown != 0 {
				return nil
			}
		}
	}
}

func pauseAndExit() {
	fmt.Println("Press any key to continue...")
	if err := waitForKeyPress(); err != nil {
		log.Errorf("Error waiting for key press: %v", err)
	}
}

func pauseAndReturnToMainMenu() {
	if _, err := fmt.Println("Press any key to return to the main menu..."); err != nil {
		log.Errorf("Error displaying message: %v", err)
	}

	if err := waitForKeyPress(); err != nil {
		log.Errorf("Error waiting for key press: %v", err)
	}

	clearTerminal()
	showMainMenu()
}

func returnToMainMenu() {
	clearTerminal()
	showMainMenu()
}

func clearTerminal() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Errorf("Failed to clear terminal: %v", err)
	}
}
