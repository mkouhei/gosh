package main

import (
	"os"
	"testing"
	"time"
)

func TestRead(t *testing.T) {
	e := NewEnv(false)
	f, err := os.OpenFile("dummy_code", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	text := `import "fmt"

func main() {
     fmt.Println("hello")
}
`
	f.WriteString(text)

	f, err = os.Open("dummy_code")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	e.shell(f)

	time.Sleep(time.Nanosecond)

	os.Remove("dummy_code")
}
