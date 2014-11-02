package main

import (
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
