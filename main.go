package main

import (
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command("go", "run", "./cmd/app")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	_ = cmd.Run()
}
