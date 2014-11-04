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

	lines := []string{"import \"fmt\"\n",
		"func main() {\n",
		"msg := \"hello\"\n"}
	if err := e.write(lines); err != nil {
		t.Fatal(err)
	}

	var lines2 = []string{"fmt.Println(msg)\n", "}\n"}
	if err := e.write(lines2); err != nil {
		t.Fatal(err)
	}
	cleanDirs(e.BldDir)
}
