package main

import (
	"fmt"
	"os"
	"testing"
	"time"
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
	f, err := os.OpenFile(tmpFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		t.Fatal("creating test file faild: %s", err)
	}
	f.Sync()

	time.Sleep(time.Microsecond)

	f.WriteString("package main")
	f.Sync()
	f.Close()

	f, err = os.OpenFile(tmpFile, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		t.Fatal("open test file failed: %s", err)
	}
	time.Sleep(time.Microsecond)

	fmt.Println("[test modify] ")
	var content = `
import "fmt"

func main() {
msg := "hello"
`

	f.WriteString(content)
	f.Sync()
	f.Close()

	f, err = os.OpenFile(tmpFile, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		t.Fatal(err)
	}

	var content2 = `
fmt.Println(msg)
}
`
	f.WriteString(content2)
	f.Sync()
	f.Close()
}
