package main

import (
	"io/ioutil"
	"os"
	"testing"
)

const (
	dirPerm  = 0755
	testDir  = "_test"
	tempFile = "_test/tempcode"
)

func cleanDirs(targetDir string) error {
	if err := os.RemoveAll(targetDir); err != nil {
		return err
	}
	return nil
}

func TestWatch(t *testing.T) {

	if err := cleanDirs(testDir); err != nil {
		t.Fatal(err)
	}

	if err := os.Mkdir(testDir, dirPerm); err != nil {
		t.Fatal(err)
	}
	go func() {
		if err := watch(testDir); err != nil {
			t.Fatal(err)
		}
	}()

	err := ioutil.WriteFile(tempFile, []byte{}, 0644)
	if err != nil {
		t.Fatal(err)
	}
}
