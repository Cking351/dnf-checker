package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	updates := checkUpdates()

	// Check if there are any updates..
	if len(strings.TrimSpace(updates)) > 0 {
		sendNotification(fmt.Sprintf("Updates available:\n%s", updates))
	} else {
		// exit and do not send notification
		return
	}
}

func checkUpdates() string {
	cmd := exec.Command("dnf", "check-upgrade")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return out.String()
}

func sendNotification(msg string) {
	cmd := exec.Command("notify-send", "DNF Update Check", msg)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
