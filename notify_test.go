package main

import (
	"fmt"
	"testing"
)

func TestWatch(t *testing.T) {

	blddir := bldDir()
	tmpFile := fmt.Sprintf("%s/%s", blddir, "gosh_tmp.go")

	go func() {
		if err := watch(blddir); err != nil {
			t.Fatal(err)
		}
	}()

	fmt.Println("[test create] ")
	if err := writeFile(tmpFile, "package main"); err != nil {
		t.Fatal(err)
	}

	fmt.Println("[test modify] ")

	var content = `
import "fmt"

func main() {
msg := "hello"
`

	if err := writeFile(tmpFile, content); err != nil {
		t.Fatal(err)
	}

	var content2 = `
fmt.Println(msg)
}
`
	if err := writeFile(tmpFile, content2); err != nil {
		t.Fatal(err)
	}
}
