package main

import (
	"os"
	"testing"
)

func TestRunCmd(t *testing.T) {
	cmd := "foo"
	args := []string{}
	if err := runCmd(cmd, args...); err == nil {
		t.Fatal("want: <fail>")
	}
	cmd = "true"
	if err := runCmd(cmd, args...); err != nil {
		t.Fatal(err)
	}
}

func TestBldDirAndCleanDir(t *testing.T) {
	d := bldDir()
	f, err := os.Stat(d)
	if err != nil {
		t.Fatal(err)
	}
	if !f.IsDir() {
		t.Fatalf("expecting directory: %s", d)
	}
	if err := cleanDir(d); err != nil {
		t.Fatal(err)
	}
}
