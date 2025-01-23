package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func main() {
	packageCount, err := checkUpdates()
	if err != nil {
		log.Fatalf("Update check failed: %v", err)
	}

	if packageCount > 0 {
		// Notify the user with the package count
		sendNotification(fmt.Sprintf("There are %d package updates available.", packageCount))
	} else {
		// No updates, exit quietly
		fmt.Println("System is up-to-date.")
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
	err := cmd.Run()

	// Check for errors and handle known exit codes
	if err != nil {
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
		}

		// For other errors, return a detailed error message
		return 0, fmt.Errorf("command failed with exit code %d: %s, error: %v", exitCode, stderr.String(), err)
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
