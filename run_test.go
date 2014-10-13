package main

import (
	"testing"
)

func TestRunCmd(t *testing.T) {
	if c := runCmd("test"); c != "test" {
		t.Fatalf("%v, want tst", c)
	}
}
