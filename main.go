package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const stateFile = "/var/tmp/update-notifier.state"

func main() {
	err := ensureStateFileExists()
	if err != nil {
		log.Fatalf("Failed to ensure state directory exists: %v", err)
	}

	packageCount, err := checkUpdates()
	if err != nil {
		log.Fatalf("Update check failed: %v", err)
	}

	if packageCount > 0 {
		shouldNotify, err := shouldSendNotification(packageCount)
		if err != nil {
			log.Fatalf("Failed to check notification state: %v", err)
		}

		if shouldNotify {
			sendNotification(fmt.Sprintf("%d package(s) available", packageCount))

			//Update state
			err = updateState(packageCount)
			if err != nil {
				log.Fatalf("Failed to update state: %v", err)
			}
		}
	} else {
		err := clearState()
		if err != nil {
			log.Fatalf("Failed to clear state: %v", err)
		}
	}
}

func checkUpdates() (int, error) {
	cmd := exec.Command("dnf", "check-upgrade")
	var out bytes.Buffer
	var stderr bytes.Buffer

	// Capture both stdout and stderr
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Run the command
	exit := cmd.Run()

	// Check for errors and handle known exit codes
	if exit != nil {
		// Extract the exit code and handle updates
		exitCode := cmd.ProcessState.ExitCode()

		// Exit code 100 means updates are available; process the output
		if exitCode == 100 {
			// Filter and count package lines
			lines := strings.Split(out.String(), "\n")
			packageCount := 0
			for _, line := range lines {
				// Ignore any blank lines or headers, count package lines only
				if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "Last metadata") && !strings.HasPrefix(line, "Obsoleting package") {
					packageCount++
				}
			}
			return packageCount, nil
		} else if exitCode == 1 {
			return 0, fmt.Errorf("error running DNF check-upgrade: %s", exit)
		}
	}

	// Exit code 0 means no updates are available
	return 0, nil
}

func sendNotification(msg string) {
	cmd := exec.Command("notify-send", "DNF Update Check", msg)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func ensureStateFileExists() error {
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		file, err := os.Create(stateFile)
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := file.Close(); closeErr != nil {
				log.Fatalf(
					"Failed to close state file: %v",
					closeErr)
			}
		}()

		_, err = file.WriteString("0")
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func shouldSendNotification(currentCount int) (bool, error) {
	if currentCount == 0 {
		return false, nil
	}

	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}

	storedCount, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return false, err
	}
	return currentCount != storedCount, nil
}

func updateState(packageCount int) error {
	return os.WriteFile(stateFile, []byte(strconv.Itoa(packageCount)), 0644)
}

func clearState() error {
	if err := os.Remove(stateFile); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
