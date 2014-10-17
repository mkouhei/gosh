package main

import (
	"io/ioutil"
	"log"
	"os"
)

const (
	dirPerm = 0755
)

func cleanDirs(targetDir string) error {
	if err := os.RemoveAll(targetDir); err != nil {
		return err
	}
	return nil
}

func bldDir() string {
	f, err := ioutil.TempDir("", "gosh-")
	if err != nil {
		log.Fatal(err)
	}
	return f
}
