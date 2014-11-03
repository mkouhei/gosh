package main

import (
	"fmt"
	"testing"
)

func TestWatch(t *testing.T) {
	e := NewEnv(false)

	go func() {
		if err := e.watch(); err != nil {
			t.Fatal(err)
		}
	}()

	fmt.Println("[test create] ")
	if err := e.initFile(); err != nil {
		t.Fatal(err)
	}

	fmt.Println("[test modify] ")

	var content = `
import "fmt"

func main() {
msg := "hello"
`

	if err := e.write(content); err != nil {
		t.Fatal(err)
	}

	var content2 = `
fmt.Println(msg)
}
`
	if err := e.write(content2); err != nil {
		t.Fatal(err)
	}
	cleanDirs(e.BldDir)
}
