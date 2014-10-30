package main

import (
	"os"
)

func setGOPATH(p string) {
	os.Setenv("GOPATH", p)
}

func goGet(p string) error {
	cmd := "go"
	args := []string{"get", p}
	if err := runCmd(cmd, args...); err != nil {
		return err
	}
	return nil
}
