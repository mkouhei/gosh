package main

import (
	"io/ioutil"
	"log"
	"os"
)

func cleanDirs(targetDir string) error {
	if err := os.RemoveAll(targetDir); err != nil {
		return err
	}
	return nil
}

func bldDir() string {
	f, err := ioutil.TempDir("", prefix)
	if err != nil {
		log.Fatal(err)
	}
	return f
}
