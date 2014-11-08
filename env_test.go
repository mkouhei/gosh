package main

import (
	"fmt"
	"os"
	"testing"
)

func TestNewEnv(t *testing.T) {
	e := NewEnv(false)
	_, err := os.Stat(e.BldDir)
	if err != nil {
		t.Fatal(err)
	}
	if e.TmpPath != fmt.Sprintf("%s/%s", e.BldDir, tmpname) {
		t.Fatal("error initialize")
	}
	if e.GoPath != e.BldDir {
		t.Fatal("error initialize")
	}
	if e.Debug {
		t.Fatal("error initialize")
	}
	cleanDir(e.BldDir)
}
