package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestRunCmd(t *testing.T) {
	cmd := "foo"
	args := []string{}
	if msg, err := runCmd(true, cmd, args...); err == nil {
		t.Fatal("want: <fail>: %s", msg)
	}
	cmd = "true"
	if msg, err := runCmd(true, cmd, args...); err != nil {
		t.Fatal(msg)
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
	cleanDirs()
	lists, _ := filepath.Glob(fmt.Sprintf("/tmp/%s*", prefix))
	if lists != nil {
		t.Fatal("FAILURE cleanDirs()")
	}
}

func TestSearchString(t *testing.T) {
	list := []string{"foo", "bar", "baz"}
	if !searchString("foo", list) {
		t.Fatal("expecting true")
	}
	if searchString("hoge", list) {
		t.Fatal("expecting false")
	}
}
