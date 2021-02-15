package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

const cmdKubectl = "kubectl"

func runCommand(cmdName string, args []string) error {
	log.Printf("Running: %s %s", cmdName, strings.Join(args, " "))
	cmd := exec.Command(cmdName, args...)

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command %s: %s", cmdName, stderr.String())
	}

	log.Printf("Output: %s", stdout.String())

	return nil
}
