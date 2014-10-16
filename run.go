package main

import (
	"fmt"
	"io"
	"os/exec"
)

func runCmd(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	buf := make([]byte, 1024)
	var n int
	for {
		if n, err = stdout.Read(buf); err != nil {
			break
		}
		fmt.Print(string(buf[0:n]))
	}
	if err == io.EOF {
		err = nil
	}
	return nil
}
